package output

import (
	"image"

	"github.com/flynn-nrg/go-vfx/go-oiio/oiio"
)

var _ Output = (*OIIO)(nil)

type OIIO struct {
	fileName string
}

func NewOIIO(fileName string) (*OIIO, error) {
	return &OIIO{
		fileName: fileName,
	}, nil
}

func (o *OIIO) Write(i image.Image) error {
	return oiio.WriteImage(o.fileName, i)
}
