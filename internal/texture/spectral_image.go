package texture

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/go-oiio/oiio"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ SpectralTexture = (*SpectralImage)(nil)

// SpectralImage represents a spectral image-based texture.
// It transforms a float image into a spectral representation with wavelength buckets 5nm apart.
type SpectralImage struct {
	sizeX int
	sizeY int
	data  image.Image
	// Spectral data organized by wavelength buckets (5nm apart from 380-750nm)
	spectralData [][]float64 // [wavelength_bucket][pixel_index]
	wavelengths  []float64   // Wavelength values for each bucket
}

// NewSpectralImageFromRawData returns a new SpectralImage instance from raw float data.
func NewSpectralImageFromRawData(width int, height int, data []float64) *SpectralImage {
	img := floatimage.NewFloatNRGBA(image.Rect(0, 0, width, height), data)
	return NewSpectralImageFromImage(img)
}

// NewSpectralImageFromFile returns a new SpectralImage instance by using the supplied file path.
func NewSpectralImageFromFile(path string) (*SpectralImage, error) {
	img, err := oiio.ReadImage(path)
	if err != nil {
		return nil, err
	}

	return NewSpectralImageFromImage(img), nil
}

// NewSpectralImageFromPNG returns a new SpectralImage instance by using the supplied PNG data.
func NewSpectralImageFromPNG(r io.Reader) (*SpectralImage, error) {
	img, err := png.Decode(r)
	if err != nil {
		return nil, err
	}

	return NewSpectralImageFromImage(img), nil
}

// NewSpectralImageFromHDR returns a new SpectralImage instance by using the supplied HDR data.
func NewSpectralImageFromHDR(fileName string) (*SpectralImage, error) {
	img, err := oiio.ReadImage(fileName)
	if err != nil {
		return nil, err
	}

	return NewSpectralImageFromImage(img), nil
}

// NewSpectralImageFromImage creates a spectral image from an existing image.
func NewSpectralImageFromImage(img image.Image) *SpectralImage {
	si := &SpectralImage{
		sizeX: img.Bounds().Dx(),
		sizeY: img.Bounds().Dy(),
		data:  img,
	}

	// Initialize spectral data with CIE wavelength buckets (5nm apart from 380-750nm)
	// Use the same wavelengths as defined in the spectral package
	si.wavelengths = []float64{
		380, 385, 390, 395, 400, 405, 410, 415, 420, 425,
		430, 435, 440, 445, 450, 455, 460, 465, 470, 475,
		480, 485, 490, 495, 500, 505, 510, 515, 520, 525,
		530, 535, 540, 545, 550, 555, 560, 565, 570, 575,
		580, 585, 590, 595, 600, 605, 610, 615, 620, 625,
		630, 635, 640, 645, 650, 655, 660, 665, 670, 675,
		680, 685, 690, 695, 700, 705, 710, 715, 720, 725,
		730, 735, 740, 745, 750,
	}

	si.spectralData = make([][]float64, len(si.wavelengths))
	for i := range si.spectralData {
		si.spectralData[i] = make([]float64, si.sizeX*si.sizeY)
	}

	// Transform RGB data to spectral representation
	si.transformRGBToSpectral()

	return si
}

// transformRGBToSpectral converts RGB image data to spectral representation.
// This uses a simple transformation where each RGB channel influences
// different wavelength ranges based on typical material properties.
func (si *SpectralImage) transformRGBToSpectral() {
	for y := 0; y < si.sizeY; y++ {
		for x := 0; x < si.sizeX; x++ {
			pixelIndex := y*si.sizeX + x

			// Get RGB value at this pixel
			var r, g, b float64
			if img, ok := si.data.(*floatimage.FloatNRGBA); ok {
				pixel := img.FloatNRGBAAt(x, y)
				r, g, b = pixel.R, pixel.G, pixel.B
			} else {
				pixel := color.NRGBAModel.Convert(si.data.At(x, y)).(color.NRGBA)
				r = float64(pixel.R) / 255.0
				g = float64(pixel.G) / 255.0
				b = float64(pixel.B) / 255.0
			}

			// Transform RGB to spectral values
			// This is a simplified transformation - in practice, you might use
			// measured spectral data or more sophisticated models
			for i, wavelength := range si.wavelengths {
				spectralValue := si.rgbToSpectralValue(r, g, b, wavelength)
				si.spectralData[i][pixelIndex] = spectralValue
			}
		}
	}
}

// rgbToSpectralValue converts RGB values to a spectral value at a specific wavelength.
// This uses an improved model with better wavelength ranges and falloff characteristics:
// - Red channel influences longer wavelengths (600-750nm) with peak at 650nm
// - Green channel influences medium wavelengths (480-620nm) with peak at 550nm
// - Blue channel influences shorter wavelengths (380-520nm) with peak at 450nm
//
// Note: The red channel has a slight bias (reduced strength by 20%) to compensate for
// the fact that many PBR materials (like metals, rust, etc.) naturally have strong
// red components in their albedo textures. Without this bias, the spectral conversion
// would over-emphasize red, making materials appear too reddish compared to the RGB reference.
// This bias helps maintain color balance while still preserving the spectral characteristics
// of the materials.
func (si *SpectralImage) rgbToSpectralValue(r, g, b, wavelength float64) float64 {
	var spectralValue float64

	// Red channel contribution (600-750nm, peak at 650nm) - reduced range
	if wavelength >= 600.0 && wavelength <= 750.0 {
		// Use a Gaussian-like falloff centered at 650nm
		center := 650.0
		distance := math.Abs(wavelength - center)
		width := 50.0 // Reduced width for less red influence
		falloff := math.Exp(-(distance * distance) / (2.0 * width * width))
		redContribution := r * falloff * 0.8 // Reduced strength by 20%
		spectralValue += redContribution
	}

	// Green channel contribution (480-620nm, peak at 550nm)
	if wavelength >= 480.0 && wavelength <= 620.0 {
		// Use a Gaussian-like falloff centered at 550nm
		center := 550.0
		distance := math.Abs(wavelength - center)
		width := 50.0
		falloff := math.Exp(-(distance * distance) / (2.0 * width * width))
		greenContribution := g * falloff
		spectralValue += greenContribution
	}

	// Blue channel contribution (380-520nm, peak at 450nm)
	if wavelength >= 380.0 && wavelength <= 520.0 {
		// Use a Gaussian-like falloff centered at 450nm
		center := 450.0
		distance := math.Abs(wavelength - center)
		width := 50.0
		falloff := math.Exp(-(distance * distance) / (2.0 * width * width))
		blueContribution := b * falloff
		spectralValue += blueContribution
	}

	// For neutral colors (when r ≈ g ≈ b), ensure truly neutral response
	// by adding a small constant across all wavelengths
	if math.Abs(r-g) < 0.1 && math.Abs(g-b) < 0.1 && math.Abs(r-b) < 0.1 {
		// This is a neutral color, preserve more brightness for specular highlights
		spectralValue = math.Max(spectralValue, r*0.95) // Increased from 0.8 to 0.95
	}

	// Clamp to [0, 1]
	return math.Max(0.0, math.Min(1.0, spectralValue))
}

// Value returns the spectral value at the given UV coordinates and wavelength.
func (si *SpectralImage) Value(u float64, v float64, lambda float64, _ *vec3.Vec3Impl) float64 {
	// Convert UV to pixel coordinates
	i := int(u * float64(si.sizeX))
	j := int((1 - v) * (float64(si.sizeY) - 0.001))

	// Clamp coordinates
	if i < 0 {
		i = 0
	}
	if j < 0 {
		j = 0
	}
	if i > (si.sizeX - 1) {
		i = si.sizeX - 1
	}
	if j > (si.sizeY - 1) {
		j = si.sizeY - 1
	}

	pixelIndex := j*si.sizeX + i

	// Find the wavelength bucket that contains the requested wavelength
	wavelengthIndex := si.findWavelengthIndex(lambda)
	if wavelengthIndex < 0 || wavelengthIndex >= len(si.spectralData) {
		return 0.0
	}

	if pixelIndex < 0 || pixelIndex >= len(si.spectralData[wavelengthIndex]) {
		return 0.0
	}

	return si.spectralData[wavelengthIndex][pixelIndex]
}

// findWavelengthIndex finds the index of the wavelength bucket that contains the given wavelength.
func (si *SpectralImage) findWavelengthIndex(lambda float64) int {
	// Clamp wavelength to valid range
	if lambda < si.wavelengths[0] {
		return 0
	}
	if lambda > si.wavelengths[len(si.wavelengths)-1] {
		return len(si.wavelengths) - 1
	}

	// Find the closest wavelength bucket
	for i, wavelength := range si.wavelengths {
		if lambda <= wavelength {
			return i
		}
	}

	return len(si.wavelengths) - 1
}

// FlipY() flips the spectral image upside down.
func (si *SpectralImage) FlipY() {
	// Flip the underlying image data
	if im, ok := si.data.(*floatimage.FloatNRGBA); ok {
		for y := si.data.Bounds().Min.Y; y <= si.data.Bounds().Max.Y/2; y++ {
			for x := si.data.Bounds().Min.X; x <= si.data.Bounds().Max.X; x++ {
				c1 := si.data.At(x, y)
				c2 := si.data.At(x, si.data.Bounds().Max.Y-y)
				im.Set(x, y, c2)
				im.Set(x, si.data.Bounds().Max.Y-y, c1)
			}
		}
	}

	// Re-transform the spectral data
	si.transformRGBToSpectral()
}

// FlipX() flips the spectral image from left to right.
func (si *SpectralImage) FlipX() {
	// Flip the underlying image data
	if im, ok := si.data.(*floatimage.FloatNRGBA); ok {
		for y := si.data.Bounds().Min.Y; y <= si.data.Bounds().Max.Y; y++ {
			for x := si.data.Bounds().Min.X; x <= si.data.Bounds().Max.X/2; x++ {
				c1 := si.data.At(x, y)
				c2 := si.data.At(si.data.Bounds().Max.X-x, y)
				im.Set(x, y, c2)
				im.Set(si.data.Bounds().Max.X-x, y, c1)
			}
		}
	}

	// Re-transform the spectral data
	si.transformRGBToSpectral()
}

// SizeX returns the width of the underlying image.
func (si *SpectralImage) SizeX() int {
	return si.sizeX
}

// SizeY returns the height of the underlying image.
func (si *SpectralImage) SizeY() int {
	return si.sizeY
}

// GetData returns the underlying image data.
func (si *SpectralImage) GetData() image.Image {
	return si.data
}

// GetWavelengths returns the wavelength array.
func (si *SpectralImage) GetWavelengths() []float64 {
	return si.wavelengths
}
