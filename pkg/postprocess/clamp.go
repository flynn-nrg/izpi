package postprocess

import (
	"errors"
	"image"

	"gitlab.com/flynn-nrg/izpi/pkg/camera"
	"gitlab.com/flynn-nrg/izpi/pkg/colour"
	"gitlab.com/flynn-nrg/izpi/pkg/floatimage"
)

// Ensure interface compliance.
var _ Filter = (*Clamp)(nil)

// Clamp represents a clamp filter.
type Clamp struct {
	max float64
}

// NewClamp returns a new clamp filter.
func NewClamp(max float64) *Clamp {
	return &Clamp{
		max: max,
	}
}

func (c *Clamp) Apply(i image.Image, cam *camera.Camera) error {
	im, ok := i.(*floatimage.FloatNRGBA)
	if !ok {
		return errors.New("only FloatNRGBA image format is supported")
	}
	for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
			pixel := im.FloatNRGBAAt(x, y)
			im.Set(x, y, colour.FloatNRGBA{
				R: clamp(pixel.R, c.max),
				G: clamp(pixel.G, c.max),
				B: clamp(pixel.B, c.max),
				A: 1.0})
		}
	}

	return nil
}

func clamp(v float64, max float64) float64 {
	if v <= max {
		return v
	}
	return max
}
