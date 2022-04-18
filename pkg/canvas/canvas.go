package canvas

import (
	"fmt"
	"io"
)

// Ensure interface compliance.
var _ Canvas = (*CanvasImpl)(nil)

// CanvasImpl defines the structure of a canvas.
type CanvasImpl struct {
	SizeX       int
	SizeY       int
	PixelSize   int
	PixelFormat int
	Palette     []int32
	Buffer      []byte
}

// New returns a new CanvasImpl.
func New(sizeX int, sizeY int, pixelFormat int, palette []byte) (*CanvasImpl, error) {
	c := &CanvasImpl{
		SizeX:       sizeX,
		SizeY:       sizeY,
		PixelFormat: pixelFormat,
	}

	switch pixelFormat {
	case PixelFormatBW:
		c.PixelSize = 1
	case PixelFormatIndexed:
		c.PixelSize = 1
	case PixelFormatRGB:
		c.PixelSize = 3
	case PixelFormatRGBA, PixelFormatARGB:
		c.PixelSize = 4
	default:
		return nil, fmt.Errorf("Unsupported pixel format %v", pixelFormat)
	}

	c.Buffer = make([]byte, sizeX*sizeY*c.PixelSize)
	return c, nil
}

// Write exports the image data to a suitable format.
func (c *CanvasImpl) Write(w io.Writer) error {
	return nil
}
