package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"

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

func deltaReport(orig *[]byte, edited *[]byte) string {
	origMeta, _ := bimg.Metadata(*orig)
	editedMeta, _ := bimg.Metadata(*edited)

	return fmt.Sprintf("Original: %+v\nNew: %+v", origMeta.Size, editedMeta.Size)
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
	log.Printf("Original metadata: %+v", srcMeta)
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
	fmt.Println(deltaReport(&srcFile, &out))
	if !opt.noAction {
		bimg.Write(opt.outputFilename, out)
		return
	}
}
