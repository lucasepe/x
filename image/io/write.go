package io

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

type Format int

const (
	JPG Format = iota + 1
	PNG
)

// WriteToFile writes the given image to a file on disk using the specified format.
//
// It creates or truncates the target file, writes the image data in the
// requested format (JPG or PNG), and then closes the file.
func WriteToFile(img image.Image, path string, as Format) (err error) {
	wri, err := os.Create(path)
	if err != nil {
		return err
	}
	defer wri.Close()

	return Write(img, wri, as)
}

// Write writes the image to the specified writer
// in the requested format (JPG or PNG).
func Write(img image.Image, wri io.Writer, as Format) (err error) {
	switch as {
	case JPG:
		err = jpeg.Encode(wri, img, &jpeg.Options{Quality: 85})
	case PNG:
		err = png.Encode(wri, img)
	default:
		err = fmt.Errorf("unsupported format: %v", as)
	}

	return err
}
