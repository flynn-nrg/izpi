package postprocess

import (
	"errors"
	"image"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/go-vfx/math32"
	"github.com/flynn-nrg/izpi/internal/scene"
)

// Ensure interface compliance.
var _ Filter = (*Gamma)(nil)

// Gamma represents a gamma adjustment filter.
type Gamma struct{}

// NewClamp returns a new clamp filter.
func NewGamma() *Gamma {
	return &Gamma{}
}

func (g *Gamma) Apply(i image.Image, _ *scene.Scene) error {
	im, ok := i.(*floatimage.Float32NRGBA)
	if !ok {
		return errors.New("only Float32NRGBA image format is supported")
	}
	for y := i.Bounds().Min.Y; y <= i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x <= i.Bounds().Max.X; x++ {
			pixel := im.Float32NRGBAAt(x, y)
			im.Set(x, y, colour.Float32NRGBA{
				R: math32.Sqrt(pixel.R),
				G: math32.Sqrt(pixel.G),
				B: math32.Sqrt(pixel.B),
				A: pixel.A})
		}
	}

	return nil
}
