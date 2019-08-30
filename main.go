package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/h2non/bimg.v1"
)

type options struct {
	version        bool
	verbosity      int
	inputFilename  string
	outputFilename string
	outputSuffix   string
	jpegQuality    int
	outputWidth    int
	outputHeight   int
	maxLongest     int
	minShortest    int
	pctResize      float64
	stretch        bool
	force          bool
	noAction       bool
	test           bool
	positionalArgs []string
}

const binName string = "goimg"
const defaultQuality int = 85

var opt options
var imgOpt bimg.Options
var termWidth, termHeight int
var log *zap.SugaredLogger

// Max calculates the maximum of two integers
func Max(nums ...int) int {
	max := nums[0]
	for _, i := range nums[1:] {
		if i > max {
			max = i
		}
	}
	return max
}

// Min calculates the minimum of two integers
func Min(nums ...int) int {
	min := nums[0]
	for _, i := range nums[1:] {
		if i < min {
			min = i
		}
	}
	return min
}

// Scale calculates the new pixel size based on pct scaling factor
func Scale(pct float64, size int) int {
	return int(float64(size) * (float64(pct) / float64(100)))
}

// Humanize prints bytes in human-readable strings
func Humanize(bytes int) string {
	suffix := "B"
	num := float64(bytes)
	factor := 1024.0
	// k=kilo, M=mega, G=giga, T=tera, P=peta, E=exa, Z=zetta, Y=yotta
	units := []string{"", "K", "M", "G", "T", "P", "E", "Z"}

	for _, unit := range units {
		if num < factor {
			return fmt.Sprintf("%3.1f%s%s", num, unit, suffix)
		}
		num = (num / factor)
	}
	// if we got here, it's a really big number!
	// return yottabytes
	return fmt.Sprintf("%.1f%s%s", num, "Y", suffix)
}

func init() {
	var err error
	if termWidth, termHeight, err = terminal.GetSize(1); err != nil {
		panic(err)
	}
	logTimeEnc := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		if termWidth > 200 {
			enc.AppendString(t.Format("Jan 02 15:04:05.000"))
		}
	}
	// init log
	atom := zap.NewAtomicLevel()
	config := zap.Config{
		Encoding:         "console",
		Level:            atom,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: logTimeEnc,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	logger, _ := config.Build()
	log = logger.Sugar()
	defer log.Sync()
	// set flags
	flag.CommandLine.SetOutput(os.Stderr)
	// Literal usage message {{{
	//     flag.Usage = func() {
	//         usage := fmt.Sprintf(
	//             `Usage: %s [flags|options] [input_file] <output_file>
	// Flags:
	//     -d      Print debug messages to console
	//     -f 	    Force overwrite output file, even if it exists
	//     -n 	    Don't write files; just display results
	// Options:
	//     -q      Quality of jpeg file (1-100; default %d) <int>
	//     -h 	    Height in pixels of output file <int>
	//     -w 	    Width in pixels of output file <int>
	//     -mh	    Maximum height of output file <int>
	//     -mw	    Maximum width of output file <int>
	//     -max    Maximum pixel size of either dimension <int>
	//     -min    Minimum pixel size of either dimension <int>
	//     -pct    Resize to pct of original dimensions <float>`, binName, defaultQuality)
	//         fmt.Fprintln(os.Stderr, usage)
	//     }
	// }}}

	flag.StringVar(&opt.inputFilename, "i", "", "path of input file")
	flag.StringVar(&opt.outputFilename, "o", "", "path of output file")
	flag.StringVar(&opt.outputSuffix, "suffix", "_edit", "suffix to add to input filename for output filename")
	flag.IntVar(&opt.jpegQuality, "q", defaultQuality, "jpeg quality (1-100)")
	flag.IntVar(&opt.outputWidth, "w", 0, "width of output file")
	flag.IntVar(&opt.outputHeight, "h", 0, "height of output file")
	flag.IntVar(&opt.maxLongest, "l", 0, "maximum length of longest dimension")
	flag.IntVar(&opt.minShortest, "s", 0, "Minimum length of shortest dimension")
	flag.Float64Var(&opt.pctResize, "p", 0, "resize to pct of original dimensions")
	flag.BoolVar(&opt.force, "f", false, "overwrite output file if it exists")
	flag.IntVar(&opt.verbosity, "v", 0, "increase debug messages to console")
	flag.BoolVar(&opt.noAction, "n", false, "don't write files; just display results")
	flag.BoolVar(&opt.test, "test", false, "use test files")
	flag.Parse()

	// Set logging level
	switch opt.verbosity {
	case 1:
		atom.SetLevel(zap.InfoLevel)
	case 2:
		atom.SetLevel(zap.DebugLevel)
	default:
		atom.SetLevel(zap.ErrorLevel)
	}
	if opt.test {
		opt.inputFilename = "./test/example01.jpg"
		opt.outputFilename = "./test/example01_edit.jpg"
	}
	if opt.inputFilename == "" {
		fmt.Fprintln(os.Stderr, "No input filename provided, quitting.")
		flag.Usage()
		os.Exit(1)
	}
	if opt.outputFilename == "" {
		log.Info("No output filename provided; applying suffix to input filename")
		ext := filepath.Ext(opt.inputFilename)
		pathNoExt := opt.inputFilename[0 : len(opt.inputFilename)-len(ext)]
		opt.outputFilename = pathNoExt + opt.outputSuffix + ext
	}
}

// WriteDelta prints a formatted message of what has changed after editing
func WriteDelta(w io.Writer, orig *[]byte, edited *[]byte) error {
	var origMeta bimg.ImageMetadata
	var editedMeta bimg.ImageMetadata
	var err error
	if origMeta, err = bimg.Metadata(*orig); err != nil {
		return err
	}
	if editedMeta, err = bimg.Metadata(*edited); err != nil {
		return err
	}
	inputSize := len(*orig)
	outputSize := len(*edited)
	log.Infof("Original metadata: %+v", origMeta)
	log.Infof("Edited metadata: %+v", editedMeta)

	// Print tabulated report
	if opt.noAction {
		fmt.Fprintln(w, "***Displaying results only***")
	}
	if opt.inputFilename != opt.outputFilename {
		fmt.Fprintf(w, "File Name:\t%s\t->\t%s\n", opt.inputFilename, opt.outputFilename)
	} else {
		fmt.Fprintf(w, "File Name\t%s\n", opt.inputFilename)
	}
	if origMeta.Size != editedMeta.Size {
		fmt.Fprintf(w, "File Dimensions\t%d x %d px\t->\t%d x %d px\n",
			origMeta.Size.Width, origMeta.Size.Height, editedMeta.Size.Width, editedMeta.Size.Height)
	} else {
		// No change in dimensions; just print one set
		fmt.Fprintf(w, "File Dimensions\t%d x %d px\n",
			editedMeta.Size.Width, editedMeta.Size.Height)
	}
	fmt.Fprintf(w, "File Size\t%s\t->\t%s\n", Humanize(inputSize), Humanize(outputSize))
	fmt.Fprintf(w, "Size Reduction\t%.1f%%", 100.0-(float64(outputSize)/float64(inputSize)*100))

	return nil
}

func main() {
	log.Infof("Command line options: %+v", opt)
	// Open image
	srcFile, err := bimg.Read(opt.inputFilename)
	if err != nil {
		panic(err)
	}
	src := bimg.NewImage(srcFile)
	srcMeta, err := src.Metadata()
	if err != nil {
		panic(err)
	}
	if opt.maxLongest > 0 {
		// calculate longest dim, and assign to pctResize
		if longest := Max(srcMeta.Size.Width, srcMeta.Size.Height); longest > opt.maxLongest {
			log.Infof("Resizing to longest dimension of %d px\n", opt.maxLongest)
			opt.pctResize = (float64(opt.maxLongest) / float64(longest)) * float64(100)
		}
	}
	if opt.minShortest > 0 {
		// calculate shortest dim, and assign to pctResize
		if shortest := Min(srcMeta.Size.Width, srcMeta.Size.Height); shortest > opt.minShortest {
			log.Infof("Resizing shortest dimension to %d px\n", opt.minShortest)
			opt.pctResize = (float64(opt.minShortest) / float64(shortest)) * float64(100)
		}
	}
	opt.outputWidth = func() int {
		if opt.pctResize > 0 {
			log.Infof("Resizing to %.1f%% of original size", opt.pctResize)
			return Scale(opt.pctResize, srcMeta.Size.Width)
		}
		return opt.outputWidth
	}()
	opt.outputHeight = func() int {
		if opt.pctResize > 0 {
			return Scale(opt.pctResize, srcMeta.Size.Height)
		}
		return opt.outputHeight
	}()

	// Set image processing options
	imgOpt.Interlace = true
	imgOpt.Quality = opt.jpegQuality
	imgOpt.Width = opt.outputWidth
	imgOpt.Height = opt.outputHeight

	// Process image
	out, err := src.Process(imgOpt)
	if err != nil {
		panic(err)
	}
	log.Debugf("Image processed with options: %+v\n", imgOpt)
	// print details table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	WriteDelta(w, &srcFile, &out)
	fmt.Fprintln(w)
	w.Flush() // write details table
	if !opt.noAction {
		bimg.Write(opt.outputFilename, out)
		return
	}
}
