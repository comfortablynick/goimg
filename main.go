package main

import (
	"flag"
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
	additionalArgs []string
}

var opt Options
var debug_mode bool

func init() {
	// init log
	// TODO: add time, etc.
	log.SetPrefix("DEBUG ")
	log.SetFlags(log.Lshortfile)
	log.SetOutput(ioutil.Discard)

	// set flags
	// flag.StringVar(&opt.inputFilename, "i", "", "name of input file to resize/transcode")
	// flag.StringVar(&opt.outputFilename, "o", "", "name of output file, also determines output type")
	flag.IntVar(&opt.jpegQuality, "q", 85, "jpeg quality (1-100)")
	flag.IntVar(&opt.outputWidth, "w", 0, "width of output file")
	flag.IntVar(&opt.outputHeight, "h", 0, "height of output file")
	flag.IntVar(&opt.maxWidth, "mw", 0, "maximum width of output file")
	flag.IntVar(&opt.maxHeight, "mh", 0, "maximum height of output file")
	flag.IntVar(&opt.maxLongest, "max", 0, "maximum length of either dimension")
	flag.IntVar(&opt.minShortest, "min", 0, "Minimum length of shortest side")
	flag.Float64Var(&opt.pctResize, "pct", 0, "resize to pct of original dimensions")
	// flag.BoolVar(&opt.stretch, "stretch", false, "perform stretching resize instead of cropping")
	flag.BoolVar(&opt.force, "f", false, "overwrite output file if it exists")
	flag.BoolVar(&opt.debug, "d", false, "print debug messages to console")
	flag.BoolVar(&opt.noAction, "n", false, "don't write files; just display results")
	flag.Parse()
	opt.additionalArgs = flag.Args()

	if debug_mode || opt.debug {
		log.SetOutput(os.Stderr)
	}
}

func main() {
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
