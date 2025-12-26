package output

import (
	"image"

	"github.com/flynn-nrg/go-vfx/go-oiio/oiio"
)

var _ Output = (*OIIO)(nil)
var _ Output = (*OIIOACES)(nil)

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

// OIIOACES writes images in ACEScg color space using OIIO's ACES-aware writer.
// This should be used for spectral rendering outputs that are in ACEScg color space.
type OIIOACES struct {
	fileName string
	metadata *oiio.ACESMetadata
}

func NewOIIOACES(fileName string, metadata *oiio.ACESMetadata) (*OIIOACES, error) {
	return &OIIOACES{
		fileName: fileName,
		metadata: metadata,
	}, nil
}

func (o *OIIOACES) Write(i image.Image) error {
	return oiio.WriteImageACES(o.fileName, i, o.metadata)
}
