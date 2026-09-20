package floatimage

import (
	"image"
	"image/color"

	"github.com/flynn-nrg/izpi/internal/floatcolor"
)

// Ensure interface compliance.
var _ image.Image = (*Float32NRGBA)(nil)

const (
	numChannels = 4
)

// Float32NRGBA represents an image made up of Float32RGBA colour information.
type Float32NRGBA struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []float32
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewFloat32NRGBA returns a new Float32NRGBAImage image with the given bounds.
func NewFloat32NRGBA(r image.Rectangle, data []float32) *Float32NRGBA {
	return &Float32NRGBA{
		Pix:    data,
		Stride: numChannels * r.Dx(),
		Rect:   r,
	}
}

func (f *Float32NRGBA) At(x, y int) color.Color {
	return f.Float32NRGBAAt(x, y)
}

func (f *Float32NRGBA) Bounds() image.Rectangle {
	return f.Rect
}

func (f *Float32NRGBA) ColorModel() color.Model {
	return floatcolor.Float32NRGBAModel
}

func (f *Float32NRGBA) Float32NRGBAAt(x, y int) floatcolor.Float32NRGBA {
	if !(image.Point{x, y}.In(f.Rect)) {
		return floatcolor.Float32NRGBA{}
	}
	i := f.PixOffset(x, y)
	s := f.Pix[i : i+numChannels : i+numChannels]
	return floatcolor.Float32NRGBA{R: s[0], G: s[1], B: s[2], A: s[3]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (f *Float32NRGBA) PixOffset(x, y int) int {
	return (y-f.Rect.Min.Y)*f.Stride + (x-f.Rect.Min.X)*numChannels
}

func (f *Float32NRGBA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(f.Rect)) {
		return
	}
	i := f.PixOffset(x, y)
	c1 := floatcolor.Float32NRGBAModel.Convert(c).(floatcolor.Float32NRGBA)
	s := f.Pix[i : i+numChannels : i+numChannels]
	s[0] = c1.R
	s[1] = c1.G
	s[2] = c1.B
	s[3] = c1.A
}
