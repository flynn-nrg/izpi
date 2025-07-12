package texture

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ SpectralTexture = (*SpectralChecker)(nil)

// SpectralChecker represents a spectral checker board pattern texture.
// This texture alternates between two spectral textures based on 3D position,
// creating a checkerboard pattern where each square has different spectral properties.
type SpectralChecker struct {
	odd  SpectralTexture
	even SpectralTexture
}

// NewSpectralChecker returns a new instance of the SpectralChecker texture.
// odd and even are SpectralTexture interfaces that define the spectral response
// for the odd and even squares of the checkerboard pattern.
func NewSpectralChecker(odd SpectralTexture, even SpectralTexture) *SpectralChecker {
	return &SpectralChecker{
		odd:  odd,
		even: even,
	}
}

// Value returns the spectral response at the given position and wavelength.
// The checkerboard pattern is determined by the 3D position (p),
// and the spectral response depends on the wavelength (lambda).
func (c *SpectralChecker) Value(u float64, v float64, lambda float64, p *vec3.Vec3Impl) float64 {
	sines := math.Sin(10.0*p.X) * math.Sin(10.0*p.Y) * math.Sin(10.0*p.Z)
	if sines < 0 {
		return c.odd.Value(u, v, lambda, p)
	}

	return c.even.Value(u, v, lambda, p)
}
