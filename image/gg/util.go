// Copyright ©2022 The gg Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gg

import (
	"fmt"
	"image"
	"image/draw"
	"io/fs"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func imageToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)
	return dst
}

func fixp(x, y float64) fixed.Point26_6 {
	return fixed.Point26_6{X: fix(x), Y: fix(y)}
}

func fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(math.Round(x * 64))
}

func unfix(x fixed.Int26_6) float64 {
	const shift, mask = 6, 1<<6 - 1
	if x >= 0 {
		return float64(x>>shift) + float64(x&mask)/64
	}
	x = -x
	if x >= 0 {
		return -(float64(x>>shift) + float64(x&mask)/64)
	}
	return 0
}

// LoadFontFace is a helper function to load the specified font file with
// the specified point size. Note that the returned `font.Face` objects
// are not thread safe and cannot be used in parallel across goroutines.
// You can usually just use the Context.LoadFontFace function instead of
// this package-level function.
func LoadFontFace(path string, dpi, points float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFontFaceFromBytes(fontBytes, dpi, points)
}

// LoadFontFaceFromFS is a helper function to load the specified font file from
// the provided filesystem and path, with the specified point size.
//
// Note that the returned `font.Face` objects are not thread safe and
// cannot be used in parallel across goroutines.
// You can usually just use the Context.LoadFontFace function instead of
// this package-level function.
func LoadFontFaceFromFS(fsys fs.FS, path string, dpi, points float64) (font.Face, error) {
	if fsys == nil {
		switch {
		case filepath.IsAbs(path):
			var (
				err  error
				orig = path
				root = filepath.FromSlash("/")
			)
			path, err = filepath.Rel(root, path)
			if err != nil {
				return nil, fmt.Errorf("could not find relative path for %q from %q: %w", orig, root, err)
			}
			fsys = os.DirFS(root)
		default:
			fsys = os.DirFS(".")
		}
	}
	fontBytes, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	return LoadFontFaceFromBytes(fontBytes, dpi, points)
}

// LoadFontFace is a helper function to load the specified font with
// the specified point size. Note that the returned `font.Face` objects
// are not thread safe and cannot be used in parallel across goroutines.
// You can usually just use the Context.LoadFontFace function instead of
// this package-level function.
func LoadFontFaceFromBytes(raw []byte, dpi, points float64) (font.Face, error) {
	f, err := opentype.Parse(raw)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    points,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	return face, err
}
