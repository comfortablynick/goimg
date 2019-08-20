// Package main provides image manipulation functions
package main

import (
	"github.com/disintegration/imaging"
)

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
