package floatimage

import (
	"image"
	"image/color"

	"github.com/flynn-nrg/izpi/internal/floatcolor"
)

// Ensure interface compliance.
var _ image.Image = (*Float64NRGBA)(nil)

// Float64NRGBA represents an image made up of Float64RGBA colour information.
type Float64NRGBA struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewFloat64NRGBA returns a new Float64NRGBAImage image with the given bounds.
func NewFloat64NRGBA(r image.Rectangle, data []float64) *Float64NRGBA {
	return &Float64NRGBA{
		Pix:    data,
		Stride: numChannels * r.Dx(),
		Rect:   r,
	}
}

func (f *Float64NRGBA) At(x, y int) color.Color {
	return f.Float64NRGBAAt(x, y)
}

func (f *Float64NRGBA) Bounds() image.Rectangle {
	return f.Rect
}

func (f *Float64NRGBA) ColorModel() color.Model {
	return floatcolor.Float64NRGBAModel
}

func (f *Float64NRGBA) Float64NRGBAAt(x, y int) floatcolor.Float64NRGBA {
	if !(image.Point{x, y}.In(f.Rect)) {
		return floatcolor.Float64NRGBA{}
	}
	i := f.PixOffset(x, y)
	s := f.Pix[i : i+numChannels : i+numChannels]
	return floatcolor.Float64NRGBA{R: s[0], G: s[1], B: s[2], A: s[3]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (f *Float64NRGBA) PixOffset(x, y int) int {
	return (y-f.Rect.Min.Y)*f.Stride + (x-f.Rect.Min.X)*numChannels
}

func (f *Float64NRGBA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(f.Rect)) {
		return
	}
	i := f.PixOffset(x, y)
	c1 := floatcolor.Float64NRGBAModel.Convert(c).(floatcolor.Float64NRGBA)
	s := f.Pix[i : i+numChannels : i+numChannels]
	s[0] = c1.R
	s[1] = c1.G
	s[2] = c1.B
	s[3] = c1.A
}
