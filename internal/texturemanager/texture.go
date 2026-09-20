package texturemanager

import (
	"github.com/flynn-nrg/izpi/internal/texture"
)

type PixelDataType uint64

const (
	PixdelDataTypeUnknown PixelDataType = iota
	PixelDataTypeRGBA32
	PixelDataTypeRGBA64
	PixelDataTypeRGBAFloat32
	PixelDataTypeRGBAFloat64
)

type SerialisedTexture struct {
	Width         uint64
	Height        uint64
	PixelDataType PixelDataType
	Data          []byte
}

func (t *SerialisedTexture) ToTxtImage() *texture.ImageTxt {
	floatData, err := t.Float64Slice()
	if err != nil {
		return nil
	}
	return texture.NewFromRawData(int(t.Width), int(t.Height), floatData)
}
