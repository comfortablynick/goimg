package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"text/tabwriter"

	"gopkg.in/h2non/bimg.v1"
)

type Options struct {
	version        bool
	debug          bool
	inputFilename  string
	outputFilename string
	jpegQuality    int
	outputWidth    int
	outputHeight   int
	maxWidth       int
	maxHeight      int
	maxLongest     int
	minShortest    int
	pctResize      float64
	stretch        bool
	force          bool
	noAction       bool
	positionalArgs []string
}

const binName string = "goimg"
const debugMode bool = true
const defaultQuality int = 85

var opt Options
var imgOpt bimg.Options

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
	// init log
	// TODO: add time, etc.
	log.SetPrefix("DEBUG ")
	log.SetFlags(log.Lshortfile)
	log.SetOutput(ioutil.Discard)

	// set flags
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		usage := fmt.Sprintf(
			`Usage: %s [flags|options] [input_file] <output_file>
Flags:
    -d      Print debug messages to console
    -f 	    Force overwrite output file, even if it exists
    -n 	    Don't write files; just display results
Options:
    -q      Quality of jpeg file (1-100; default %d) <int>
    -h 	    Height in pixels of output file <int>
    -w 	    Width in pixels of output file <int>
    -mh	    Maximum height of output file <int>
    -mw	    Maximum width of output file <int>
    -max    Maximum pixel size of either dimension <int>
    -min    Minimum pixel size of either dimension <int>
    -pct    Resize to pct of original dimensions <float>`, binName, defaultQuality)
		fmt.Fprintln(os.Stderr, usage)
	}

	// flag.StringVar(&opt.inputFilename, "i", "", "name of input file to resize/transcode")
	// flag.StringVar(&opt.outputFilename, "o", "", "name of output file, also determines output type")
	// flag.BoolVar(&opt.stretch, "stretch", false, "perform stretching resize instead of cropping")
	flag.IntVar(&opt.jpegQuality, "q", defaultQuality, "jpeg quality (1-100)")
	flag.IntVar(&opt.outputWidth, "w", 0, "width of output file")
	flag.IntVar(&opt.outputHeight, "h", 0, "height of output file")
	flag.IntVar(&opt.maxWidth, "mw", 0, "maximum width of output file")
	flag.IntVar(&opt.maxHeight, "mh", 0, "maximum height of output file")
	flag.IntVar(&opt.maxLongest, "max", 0, "maximum length of either dimension")
	flag.IntVar(&opt.minShortest, "min", 0, "Minimum length of shortest side")
	flag.Float64Var(&opt.pctResize, "pct", 0, "resize to pct of original dimensions")
	flag.BoolVar(&opt.force, "f", false, "overwrite output file if it exists")
	flag.BoolVar(&opt.debug, "d", false, "print debug messages to console")
	flag.BoolVar(&opt.noAction, "n", false, "don't write files; just display results")
	flag.Parse()
	opt.positionalArgs = flag.Args()
	if debugMode || opt.debug {
		log.SetOutput(os.Stderr)
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		log.Printf("User home dir: %s", user.HomeDir)
		opt.inputFilename = "/home/nick/Dropbox/Photography/Chin Class.jpg"
		opt.outputFilename = "/home/nick/Dropbox/Photography/Chin_Class_edit.jpg"
	} else {
		opt.inputFilename = flag.Arg(1)
		opt.outputFilename = flag.Arg(2)
	}
	if opt.inputFilename == "" {
		fmt.Fprintln(os.Stderr, "No input filename provided, quitting.")
		flag.Usage()
		os.Exit(1)
	}
}

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
	log.Printf("Original metadata: %+v", origMeta)
	log.Printf("Edited metadata: %+v", editedMeta)
	// Print tabulated report
	fmt.Fprintf(w, "Input File\t%s\n", opt.inputFilename)
	fmt.Fprintf(w, "Output File\t%s\n", opt.outputFilename)
	if origMeta.Size != editedMeta.Size {
		fmt.Fprintf(w, "File Dimensions\t%d x %d px\t -> \t%d x %d px\n",
			origMeta.Size.Width, origMeta.Size.Height, editedMeta.Size.Width, editedMeta.Size.Height)
	} else {
		// No change in dimensions; just print one set
		fmt.Fprintf(w, "File Dimensions\t%d x %d px\n",
			editedMeta.Size.Width, editedMeta.Size.Height)
	}
	fmt.Fprintf(w, "File Size\t%s\t -> \t%s\n", Humanize(inputSize), Humanize(outputSize))
	fmt.Fprintf(w, "Size Reduction\t%.1f%%", 100.0-(float64(outputSize)/float64(inputSize)*100))

	return nil
}

func main() {
	opt.noAction = true
	log.Printf("Command line options: %+v", opt)
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
	if opt.noAction {
		fmt.Println("***Displaying results only***")
	}
	imgOpt.Interlace = true
	imgOpt.Width = func() int {
		if opt.outputWidth == 0 {
			if opt.outputHeight > 0 {
				return 0
			}
			return srcMeta.Size.Width
		}
		return opt.outputWidth
	}()
	imgOpt.Height = func() int {
		if opt.outputHeight == 0 {
			if opt.outputWidth > 0 {
				return 0
			}
			return srcMeta.Size.Height
		}
		return opt.outputHeight
	}()
	imgOpt.Quality = opt.jpegQuality

	// Process image
	out, err := src.Process(imgOpt)
	if err != nil {
		panic(err)
	}
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
