package postprocess

import (
	"errors"
	"image"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/izpi/internal/scene"
)

// Ensure interface compliance.
var _ Filter = (*Clamp)(nil)

// Clamp represents a clamp filter.
type Clamp struct {
	max float32
}

// NewClamp returns a new clamp filter.
func NewClamp(max float32) *Clamp {
	return &Clamp{
		max: max,
	}
}

func (c *Clamp) Apply(i image.Image, _ *scene.Scene) error {
	im, ok := i.(*floatimage.Float32NRGBA)
	if !ok {
		return errors.New("only Float32NRGBA image format is supported")
	}
	for y := i.Bounds().Min.Y; y <= i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x <= i.Bounds().Max.X; x++ {
			pixel := im.Float32NRGBAAt(x, y)
			im.Set(x, y, colour.Float32NRGBA{
				R: clamp(pixel.R, c.max),
				G: clamp(pixel.G, c.max),
				B: clamp(pixel.B, c.max),
				A: pixel.A})
		}
	}

	return nil
}

func clamp(v float32, max float32) float32 {
	if v < max {
		return v
	}
	return max
}
