// Package spectrum implements functions to work with Spectral Power Distribution
package spectrum

import (
	"github.com/flynn-nrg/izpi/pkg/common"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Spectrum defines the methods required to work with Spectral Power Distribution data.
type Spectrum interface {
	At(i int) float64
	Add(sp Spectrum) Spectrum
	Sub(sp Spectrum) Spectrum
	Mul(sp Spectrum) Spectrum
	Div(sp Spectrum) Spectrum
	Pow(p float64) Spectrum
	Exp() Spectrum
	Clamp(low, high float64) Spectrum
	Sqrt() Spectrum
	IsBlack() bool
	HasNaNs() bool
	Equal(sp Spectrum) bool
}

type SpectrumType int

const (
	Reflectance SpectrumType = iota
	Illuminant
)

func XYZToRGB(xyz *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{
		X: 3.240479*xyz.X - 1.537150*xyz.Y - 0.498535*xyz.Z,
		Y: -0.969256*xyz.X + 1.875991*xyz.Y + 0.041556*xyz.Z,
		Z: 0.055648*xyz.X - 0.204043*xyz.Y + 1.057311*xyz.Z,
	}
}

func RGBToXYZ(rgb *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{
		X: 0.412453*rgb.X + 0.357580*rgb.Y + 0.180423*rgb.Z,
		Y: 0.212671*rgb.X + 0.715160*rgb.Y + 0.072169*rgb.Z,
		Z: 0.019334*rgb.X + 0.119193*rgb.Y + 0.950227*rgb.Z,
	}
}

func findInterval(size int, lambda []float64, l float64) int {
	first := 0
	len := size
	for {
		if len <= 0 {
			break
		}
		half := len >> 1
		middle := first + half
		if lambda[middle] <= l {
			first = middle + 1
			len -= half + 1
		} else {
			len = half
		}
	}

	return common.Clamp(first-1, 0, size-1)
}

func interpolateSpectrumSamples(lambda []float64, values []float64, n int, l float64) float64 {
	if l <= lambda[0] {
		return values[0]
	}

	if l >= lambda[n-1] {
		return values[n-1]
	}

	offset := findInterval(n, lambda, l)
	t := (l - lambda[offset]) / (lambda[offset+1] - lambda[offset])
	return common.Lerp(values[offset], values[offset+1], t)
}
