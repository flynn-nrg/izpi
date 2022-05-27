package floatimage

import (
	"image"
	"image/color"

	"github.com/flynn-nrg/izpi/pkg/colour"
	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/hdrcolor"
)

// Ensure interface compliance.
var _ image.Image = (*FloatNRGBA)(nil)

// FloatNRGBA represents an image made up of FloatNRGBA colour information.
type FloatNRGBA struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []float64
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewFloatNRGBA returns a new FloatNRGBAImage image with the given bounds.
func NewFloatNRGBA(r image.Rectangle) *FloatNRGBA {
	return &FloatNRGBA{
		Pix:    make([]float64, 8*r.Dx()*r.Dy()),
		Stride: 8 * r.Dx(),
		Rect:   r,
	}
}

func (f *FloatNRGBA) At(x, y int) color.Color {
	return f.FloatNRGBAAt(x, y)
}

func (f *FloatNRGBA) Bounds() image.Rectangle {
	return f.Rect
}

func (f *FloatNRGBA) ColorModel() color.Model {
	return colour.FloatNRGBAModel
}

func (f *FloatNRGBA) FloatNRGBAAt(x, y int) colour.FloatNRGBA {
	if !(image.Point{x, y}.In(f.Rect)) {
		return colour.FloatNRGBA{}
	}
	i := f.PixOffset(x, y)
	s := f.Pix[i : i+8 : i+8]
	return colour.FloatNRGBA{R: s[0], G: s[1], B: s[2], A: s[3]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (f *FloatNRGBA) PixOffset(x, y int) int {
	return (y-f.Rect.Min.Y)*f.Stride + (x-f.Rect.Min.X)*8
}

func (f *FloatNRGBA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(f.Rect)) {
		return
	}
	i := f.PixOffset(x, y)
	c1 := colour.FloatNRGBAModel.Convert(c).(colour.FloatNRGBA)
	s := f.Pix[i : i+8 : i+8]
	s[0] = c1.R
	s[1] = c1.G
	s[2] = c1.B
	s[3] = c1.A
}

func (f *FloatNRGBA) ToHDR() (hdr.Image, error) {
	rgbImage := hdr.NewRGB(f.Rect)
	for y := f.Bounds().Min.Y; y <= f.Bounds().Max.Y; y++ {
		for x := f.Bounds().Min.X; x <= f.Bounds().Max.X; x++ {
			pixel := f.FloatNRGBAAt(x, y)
			rgbImage.Set(x, y, hdrcolor.RGB{
				R: pixel.R,
				G: pixel.G,
				B: pixel.B,
			})
		}
	}
	return rgbImage, nil
}
