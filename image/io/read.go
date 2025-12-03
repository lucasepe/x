package io

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)

func ReadFromFile(path string, as Format) (img image.Image, err error) {
	fin, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	return Read(fin, as)
}

func Read(r io.Reader, as Format) (img image.Image, err error) {
	switch as {
	case JPG:
		img, err = jpeg.Decode(r)
	case PNG:
		img, err = png.Decode(r)
	default:
		err = fmt.Errorf("unsupported format: %v", as)
	}

	return
}
