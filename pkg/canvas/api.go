// Package canvas implements a frame buffer and its associated methods.
package canvas

import "io"

const (
	// PixelFormatInvalid is the invalid pixel format.
	PixelFormatInvalid int = iota
	// PixelFormatBW is the 1bpp black and white image format.
	PixelFormatBW
	// PixelFormatIndexed is the 8bpp indexed image format.
	PixelFormatIndexed
	// PixelFormatRGB is the 24bpp RGB pixel image format.
	PixelFormatRGB
	// PixelFormatARGB is the 32bpp ARGB pixel image format.
	PixelFormatARGB
	// PixelFormatRGBA is the 32bpp RGBA pixel image format.
	PixelFormatRGBA
)

// Canvas defines the methods for the canvas operations.
type Canvas interface {
	Write(w io.Writer) error
}
