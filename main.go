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
    -pct    Resize to pct of original dimensions <float>
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
	if debugMode || opt.debug {
		log.SetOutput(os.Stderr)
	}
	opt.positionalArgs = flag.Args()
	opt.inputFilename = func() string {
		if debugMode {
			return "/home/nick/Dropbox/Photography/Chin Class.jpg"
		}
		return flag.Arg(1)
	}()
	opt.outputFilename = func() string {
		if debugMode {
			return "/home/nick/Dropbox/Photography/Chin_Class_edit.jpg"
		}
		return flag.Arg(2)
	}()
	if opt.inputFilename == "" {
		fmt.Fprintln(os.Stderr, "error: input file is required")
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	log.Printf("Command line options: %+v", opt)

	// Open image
	src, err := imaging.Open(opt.inputFilename, imaging.AutoOrientation(true))
	if err != nil {
		panic(err)
	}
	src = imaging.Resize(src, 2400, 0, imaging.Lanczos)
	err = imaging.Save(src, opt.outputFilename, imaging.JPEGQuality(75))
	if err != nil {
		panic(err)
	}
}
