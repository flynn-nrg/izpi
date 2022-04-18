package texture

import (
	"image"
	"image/color"
	"image/png"
	"io"

	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Texture = (*ImageTxt)(nil)

// ImageTxt represents an image-based texture.
type ImageTxt struct {
	sizeX      int
	sizeY      int
	colorModel color.Model
	data       image.Image
}

// NewFromPNG returns a new ImageTxt instance by using the supplied PNG data.
func NewFromPNG(r io.Reader) (*ImageTxt, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return &ImageTxt{
		sizeX:      img.Bounds().Max.X,
		sizeY:      img.Bounds().Max.Y,
		colorModel: img.Bounds().ColorModel(),
		data:       img,
	}, nil
}

func (it *ImageTxt) Value(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
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

	pixel := color.NRGBAModel.Convert(it.data.At(i, j)).(color.NRGBA)
	r := pixel.R
	g := pixel.G
	b := pixel.B
	return &vec3.Vec3Impl{X: float64(r) / 255.0, Y: float64(g) / 255.0, Z: float64(b) / 255.0}
}
