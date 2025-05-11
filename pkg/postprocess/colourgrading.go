package postprocess

import (
	"errors"
	"image"
	"io"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/gube/gube"
	"github.com/flynn-nrg/izpi/pkg/scene"
)

// Ensure interface compliance.
var _ Filter = (*ColourGrading)(nil)

// ColourGrading represents a colour grading filter.
type ColourGrading struct {
	g gube.Gube
}

// NewColourGradingFromCube returns a new colour grading filter.
func NewColourGradingFromCube(r io.Reader) (*ColourGrading, error) {
	g, err := gube.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	return &ColourGrading{
		g: g,
	}, nil
}

func (cg *ColourGrading) Apply(i image.Image, _ *scene.Scene) error {
	im, ok := i.(*floatimage.FloatNRGBA)
	if !ok {
		return errors.New("only FloatNRGBA image format is supported")
	}
	for y := i.Bounds().Min.Y; y <= i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x <= i.Bounds().Max.X; x++ {
			pixel := im.FloatNRGBAAt(x, y)
			rgb, err := cg.g.LookUp(pixel.R, pixel.G, pixel.B)
			if err != nil {
				return err
			}
			im.Set(x, y, colour.FloatNRGBA{
				R: rgb[0],
				G: rgb[1],
				B: rgb[2],
				A: pixel.A})
		}
	}

	return nil
}
