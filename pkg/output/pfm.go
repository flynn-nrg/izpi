package output

import (
	"errors"
	"image"
	"io"
	"os"

	"github.com/mdouchement/hdr"
	"github.com/mdouchement/hdr/codec/pfm"
)

// Ensure interface compliance.
var _ Output = (*PFM)(nil)

type PFM struct {
	w io.Writer
}

// NewPFM returns a new PFM output.
func NewPFM(fileName string) (*PFM, error) {
	file, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return &PFM{
		w: file,
	}, nil
}

func (p *PFM) Write(i image.Image) error {
	if hdr, ok := i.(hdr.Image); ok {
		return pfm.Encode(p.w, hdr)
	}
	return errors.New("image format must be HDR")
}
