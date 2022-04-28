package output

import (
	"image"
	"image/png"
	"io"
	"os"
)

// Ensure interface compliance.
var _ Output = (*PNG)(nil)

type PNG struct {
	w io.Writer
}

// NewPNG returns a new PNG output.
func NewPNG(fileName string) (*PNG, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return &PNG{
		w: file,
	}, nil
}

func (p *PNG) Write(i image.Image) error {
	return png.Encode(p.w, i)
}
