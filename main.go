package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/disintegration/imaging"
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

func init() {
	// init log
	// TODO: add time, etc.
	log.SetPrefix("DEBUG ")
	log.SetFlags(log.Lshortfile)
	log.SetOutput(ioutil.Discard)

	// set flags
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`Usage: %s [flags|options] [input_file] <output_file>

Flags:
    -d 	Print debug messages to console
    -f 	Force overwrite output file, even if it exists
    -n 	Don't write files; just display results

Options:
    -q 	<int> 	Quality of jpeg file (1-100; default %d)
    -h 	<int>	Height in pixels of output file
    -w 	<int> 	Width in pixels of output file
    -mh	<int> 	Maximum height of output file
    -mw	<int> 	Maximum width of output file
    -max<int> 	Maximum pixel size of either dimension
    -min<int> 	Minimum pixel size of either dimension
    -pct<float>	Resize to pct of original dimensions
  `, binName, defaultQuality)
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
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "error: input file is required")
		flag.Usage()
		os.Exit(1)
	}
	opt.positionalArgs = flag.Args()
	opt.inputFilename = flag.Arg(1)
	opt.outputFilename = flag.Arg(2)

	if debugMode || opt.debug {
		log.SetOutput(os.Stderr)
	}
}

func main() {
	log.Printf("Command line options: %+v", opt)

	// Open image
	src, err := imaging.Open("/home/nick/Dropbox/Photography/Chin Class.jpg", imaging.AutoOrientation(true))
	if err != nil {
		panic(err)
	}
	src = imaging.Resize(src, 2400, 0, imaging.Lanczos)
	err = imaging.Save(src, "/home/nick/Dropbox/Photography/Chin_Class_edit.jpg", imaging.JPEGQuality(75))
	if err != nil {
		panic(err)
	}
}
