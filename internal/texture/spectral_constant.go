package texture

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/spectral"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ SpectralTexture = (*SpectralConstant)(nil)

// SpectralConstant represents a spectral texture with either Gaussian or tabulated response.
// This texture provides wavelength-dependent responses for realistic colored materials.
type SpectralConstant struct {
	peakValue        float32                             // Maximum reflectance at the center wavelength (for Gaussian)
	centerWavelength float32                             // Wavelength where response is maximum (for Gaussian)
	width            float32                             // Width of the Gaussian curve (for Gaussian)
	spd              *spectral.SpectralPowerDistribution // Tabulated spectral response
	useTabulated     bool                                // Whether to use tabulated data or Gaussian
}

// NewSpectralConstant returns a new spectral constant texture with Gaussian response.
// peakValue: maximum reflectance (0.0 to 1.0)
// centerWavelength: wavelength where response is maximum (380-750nm)
// width: width of the response curve (smaller = narrower, larger = broader)
func NewSpectralConstant(peakValue, centerWavelength, width float32) *SpectralConstant {
	return &SpectralConstant{
		peakValue:        peakValue,
		centerWavelength: centerWavelength,
		width:            width,
		useTabulated:     false,
	}
}

// NewSpectralConstantFromSPD returns a new spectral constant texture with tabulated response.
// spd: SpectralPowerDistribution containing exact wavelength-response pairs
func NewSpectralConstantFromSPD(spd *spectral.SpectralPowerDistribution) *SpectralConstant {
	return &SpectralConstant{
		spd:          spd,
		useTabulated: true,
	}
}

// NewSpectralNeutral returns a spectral texture for neutral materials (white, gray, black).
// This creates a truly neutral response with constant reflectance across all wavelengths.
func NewSpectralNeutral(reflectance float32) *SpectralConstant {
	// Create a tabulated SPD with constant value across all wavelengths
	wavelengths := []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750}
	values := make([]float32, len(wavelengths))
	for i := range values {
		values[i] = reflectance
	}

	spd := spectral.NewSPD(wavelengths, values)
	return &SpectralConstant{
		spd:          spd,
		useTabulated: true,
	}
}

// Value returns the spectral response at the given wavelength.
// For tabulated data: interpolates between the nearest wavelength values
// For Gaussian: uses the Gaussian function centered at centerWavelength
func (c *SpectralConstant) Value(_ float32, _ float32, lambda float32, _ *vec3.Vec3Impl) float32 {
	if c.useTabulated {
		return c.interpolateSPD(lambda)
	}

	// Gaussian response: peakValue * exp(-((lambda - centerWavelength) / width)^2)
	exponent := -math.Pow(float64((lambda-c.centerWavelength)/c.width), 2)
	return c.peakValue * float32(math.Exp(exponent))
}

// interpolateSPD interpolates the spectral response at the given wavelength
// using the tabulated SpectralPowerDistribution data
func (c *SpectralConstant) interpolateSPD(lambda float32) float32 {
	if c.spd == nil || len(c.spd.Wavelengths()) == 0 {
		return 0.0
	}

	// Clamp wavelength to valid range
	if lambda < c.spd.Wavelengths()[0] {
		return c.spd.Values()[0]
	}
	if lambda > c.spd.Wavelengths()[len(c.spd.Wavelengths())-1] {
		return c.spd.Values()[len(c.spd.Values())-1]
	}

	// Find the two wavelengths that bracket the input lambda
	for i := 0; i < len(c.spd.Wavelengths())-1; i++ {
		w1 := c.spd.Wavelengths()[i]
		w2 := c.spd.Wavelengths()[i+1]

		if lambda >= w1 && lambda <= w2 {
			// Linear interpolation between the two points
			t := (lambda - w1) / (w2 - w1)
			v1 := c.spd.Values()[i]
			v2 := c.spd.Values()[i+1]
			return v1 + t*(v2-v1)
		}
	}

	// If we get here, lambda is outside the range
	return 0.0
}
