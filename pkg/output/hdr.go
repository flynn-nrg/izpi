package output

import (
	"errors"
	"image"
	"io"
	"os"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/codec/rgbe"
)

// Ensure interface compliance.
var _ Output = (*HDR)(nil)

type HDR struct {
	w io.Writer
}

// NewPNG returns a new PNG output.
func NewHDR(fileName string) (*HDR, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return &HDR{
		w: file,
	}, nil
}

func (p *HDR) Write(i image.Image) error {
	if hdr, ok := i.(hdr.Image); ok {
		return rgbe.Encode(p.w, hdr)
	}
	return errors.New("image format must be HDR")
}
