package texture

import (
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/go-vfx/go-oiio/oiio"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
)

// Ensure interface compliance.
var _ Texture = (*ImageTxt)(nil)

// ImageTxt represents an image-based texture.
type ImageTxt struct {
	sizeX int
	sizeY int
	data  image.Image
}

func NewFromRawData(width int, height int, data []float32) *ImageTxt {
	img := floatimage.NewFloat32NRGBA(image.Rect(0, 0, width, height), data)

	return &ImageTxt{sizeX: width, sizeY: height, data: img}
}

// NewFromFile returns a new ImageTxt instance by using the supplied file path.
func NewFromFile(path string) (*ImageTxt, error) {
	img, err := oiio.ReadImage64(path)
	if err != nil {
		return nil, err
	}

	return &ImageTxt{
		sizeX: img.Bounds().Dx(),
		sizeY: img.Bounds().Dy(),
		data:  img,
	}, nil
}

// NewFromPNG returns a new ImageTxt instance by using the supplied PNG data.
func NewFromPNG(r io.Reader) (*ImageTxt, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return &ImageTxt{
		sizeX: img.Bounds().Dx(),
		sizeY: img.Bounds().Dy(),
		data:  img,
	}, nil
}

// NewFromHDR returns a new ImageTxt instance by using the supplied HDR data.
func NewFromHDR(fileName string) (*ImageTxt, error) {
	img, err := oiio.ReadImage64(fileName)
	if err != nil {
		return nil, err
	}

	return &ImageTxt{
		sizeX: img.Bounds().Dx(),
		sizeY: img.Bounds().Dy(),
		data:  img,
	}, nil

}

func (it *ImageTxt) Value(u float32, v float32, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	i := int(u * float32(it.sizeX))
	j := int((1 - v) * (float32(it.sizeY) - 0.001))

	if i < 0 {
		i = 0
	}
	if j < 0 {
		j = 0
	}
	if i > (it.sizeX - 1) {
		i = it.sizeX - 1
	}

	if j > (it.sizeY - 1) {
		j = it.sizeY - 1
	}

	if img, ok := it.data.(*floatimage.Float32NRGBA); ok {
		pixel := img.Float32NRGBAAt(i, j)
		return &vec3.Vec3Impl{X: pixel.R, Y: pixel.G, Z: pixel.B}
	}

	pixel := color.NRGBAModel.Convert(it.data.At(i, j)).(color.NRGBA)
	r := pixel.R
	g := pixel.G
	b := pixel.B
	return &vec3.Vec3Impl{X: float32(r) / 255.0, Y: float32(g) / 255.0, Z: float32(b) / 255.0}
}

// FlipY() flips the image upside down.
func (it *ImageTxt) FlipY() {
	im, ok := it.data.(*floatimage.Float32NRGBA)
	if !ok {
		return
	}
	for y := it.data.Bounds().Min.Y; y <= it.data.Bounds().Max.Y/2; y++ {
		for x := it.data.Bounds().Min.X; x <= it.data.Bounds().Max.X; x++ {
			c1 := it.data.At(x, y)
			c2 := it.data.At(x, it.data.Bounds().Max.Y-y)
			im.Set(x, y, c2)
			im.Set(x, it.data.Bounds().Max.Y-y, c1)
		}
	}
}

// FlipX() flips the image from left to right.
func (it *ImageTxt) FlipX() {
	im, ok := it.data.(*floatimage.Float32NRGBA)
	if !ok {
		return
	}
	for y := it.data.Bounds().Min.Y; y <= it.data.Bounds().Max.Y; y++ {
		for x := it.data.Bounds().Min.X; x <= it.data.Bounds().Max.X/2; x++ {
			c1 := it.data.At(x, y)
			c2 := it.data.At(it.data.Bounds().Max.X-x, y)
			im.Set(x, y, c2)
			im.Set(it.data.Bounds().Max.X-x, y, c1)
		}
	}
}

// SizeX returns the width of the underlying image.
func (it *ImageTxt) SizeX() int {
	return it.sizeX
}

// SizeX returns the height of the underlying image.
func (it *ImageTxt) SizeY() int {
	return it.sizeY
}

// GetData returns the underlying image data.
func (it *ImageTxt) GetData() image.Image {
	return it.data
}
