package postprocess

import (
	"errors"
	"image"
	"math"

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
	im, ok := i.(*floatimage.float32NRGBA)
	if !ok {
		return errors.New("only float32NRGBA image format is supported")
	}
	for y := i.Bounds().Min.Y; y <= i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x <= i.Bounds().Max.X; x++ {
			pixel := im.float32NRGBAAt(x, y)
			im.Set(x, y, colour.float32NRGBA{
				R: math.Sqrt(pixel.R),
				G: math.Sqrt(pixel.G),
				B: math.Sqrt(pixel.B),
				A: pixel.A})
		}
	}

	return nil
}
