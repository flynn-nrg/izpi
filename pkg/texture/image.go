package texture

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/flynn-nrg/izpi/pkg/vec3"
	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"
	"github.com/mdouchement/hdr/hdrcolor"
)

// Ensure interface compliance.
var _ Texture = (*ImageTxt)(nil)

// ImageTxt represents an image-based texture.
type ImageTxt struct {
	sizeX int
	sizeY int
	data  image.Image
}

// NewFromPNG returns a new ImageTxt instance by using the supplied PNG data.
func NewFromPNG(r io.Reader) (*ImageTxt, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return &ImageTxt{
		sizeX: img.Bounds().Dx(),
		sizeY: img.Bounds().Dy(),
		data:  img,
	}, nil
}

// NewFromHDR returns a new ImageTxt instance by using the supplied HDR data.
func NewFromHDR(r io.Reader) (*ImageTxt, error) {
	m, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	if img, ok := m.(hdr.Image); ok {
		return &ImageTxt{
			sizeX: img.Bounds().Dx(),
			sizeY: img.Bounds().Dy(),
			data:  img,
		}, nil
	}

	return nil, errors.New("not an HDR image")
}

func (it *ImageTxt) Value(u float64, v float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	i := int(u * float64(it.sizeX))
	j := int((1 - v) * (float64(it.sizeY) - 0.001))

	if i < 0 {
		i = 0
	}
	if j < 0 {
		j = 0
	}
	if i > (it.sizeX - 1) {
		i = it.sizeX - 1
	}

	if j > (it.sizeY - 1) {
		j = it.sizeY - 1
	}

	if img, ok := it.data.(hdr.Image); ok {
		pixel := hdrcolor.RGBModel.Convert(img.At(i, j)).(hdrcolor.RGB)
		return &vec3.Vec3Impl{X: pixel.R, Y: pixel.G, Z: pixel.B}
	}

	pixel := color.NRGBAModel.Convert(it.data.At(i, j)).(color.NRGBA)
	r := pixel.R
	g := pixel.G
	b := pixel.B
	return &vec3.Vec3Impl{X: float64(r) / 255.0, Y: float64(g) / 255.0, Z: float64(b) / 255.0}
}

// FlipY() flips the image upside down.
func (it *ImageTxt) FlipY() {
	im, ok := it.data.(hdr.ImageSet)
	if !ok {
		return
	}
	for y := it.data.Bounds().Min.Y; y <= it.data.Bounds().Max.Y/2; y++ {
		for x := it.data.Bounds().Min.X; x <= it.data.Bounds().Max.X; x++ {
			c1 := it.data.At(x, y)
			c2 := it.data.At(x, it.data.Bounds().Max.Y-y)
			im.Set(x, y, c2)
			im.Set(x, it.data.Bounds().Max.Y-y, c1)
		}
	}
}

// FlipX() flips the image from left to right.
func (it *ImageTxt) FlipX() {
	im, ok := it.data.(hdr.ImageSet)
	if !ok {
		return
	}
	for y := it.data.Bounds().Min.Y; y <= it.data.Bounds().Max.Y; y++ {
		for x := it.data.Bounds().Min.X; x <= it.data.Bounds().Max.X/2; x++ {
			c1 := it.data.At(x, y)
			c2 := it.data.At(it.data.Bounds().Max.X-x, y)
			im.Set(x, y, c2)
			im.Set(it.data.Bounds().Max.X-x, y, c1)
		}
	}
}

// SizeX returns the width of the underlying image.
func (it *ImageTxt) SizeX() int {
	return it.sizeX
}

// SizeX returns the height of the underlying image.
func (it *ImageTxt) SizeY() int {
	return it.sizeY
}
