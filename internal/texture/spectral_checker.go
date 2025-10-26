package texture

import (
	"github.com/flynn-nrg/go-vfx/math32"

	"github.com/flynn-nrg/go-vfx/math32/vec3"
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
func (c *SpectralChecker) Value(u float32, v float32, lambda float32, p vec3.Vec3Impl) float32 {
	sines := math32.Sin(10.0*p.X) * math32.Sin(10.0*p.Y) * math32.Sin(10.0*p.Z)
	if sines < 0 {
		return c.odd.Value(u, v, lambda, p)
	}

	return c.even.Value(u, v, lambda, p)
}
