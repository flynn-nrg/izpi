package spectrum

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/common"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Spectrum = (*RGB)(nil)

type RGB struct {
	c            []float64
	spectrumType SpectrumType
}

func RGBFromRGB(rgb *vec3.Vec3Impl, spectrumType SpectrumType) *RGB {
	return &RGB{
		c:            []float64{rgb.X, rgb.Y, rgb.Z},
		spectrumType: spectrumType,
	}
}

func RGBFromXYZ(xyz *vec3.Vec3Impl, spectrumType SpectrumType) *RGB {
	values := XYZToRGB(xyz)
	return &RGB{
		c:            []float64{values.X, values.Y, values.Z},
		spectrumType: spectrumType,
	}
}

func RGBFromSampled(lambda, v []float64, n int) *RGB {
	xyz := &vec3.Vec3Impl{}
	for i := range cieLambda {
		val := interpolateSpectrumSamples(lambda, v, n, cieLambda[i])
		xyz.X += val * cieX[i]
		xyz.Y += val * cieY[i]
		xyz.Z += val * cieZ[i]
	}

	scale := 1.0 //float64(cieLambda[nCIESamples-1] - cieLambda[0]) / float64(CIE_Y_integral) * nCIESamples)
	xyz.X *= scale
	xyz.Y *= scale
	xyz.Z *= scale
	return RGBFromXYZ(xyz, Reflectance)
}

func (rgb *RGB) ToRGB() *vec3.Vec3Impl {
	return &vec3.Vec3Impl{
		X: rgb.c[0],
		Y: rgb.c[1],
		Z: rgb.c[2],
	}
}

func (rgb *RGB) ToRGBSpectrum() *RGB {
	return rgb
}

func (rgb *RGB) ToXYZ() *vec3.Vec3Impl {
	return RGBToXYZ(rgb.ToRGB())
}

func (rgb *RGB) Y() float64 {
	yWeight := []float64{0.212671, 0.715160, 0.072169}
	return yWeight[0]*rgb.c[0] + yWeight[1]*rgb.c[1] + yWeight[2]*rgb.c[2]
}

func (rgb *RGB) At(i int) float64 {
	return rgb.c[i]
}

func (rgb *RGB) Add(sp Spectrum) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] += sp.At(i)
	}

	return &res
}

func (rgb *RGB) Sub(sp Spectrum) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] -= sp.At(i)
	}

	return &res
}

func (rgb *RGB) Mul(sp Spectrum) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] *= sp.At(i)
	}

	return &res
}

func (rgb *RGB) ScalarMul(v float64) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] *= v
	}

	return &res
}

func (rgb *RGB) Div(sp Spectrum) Spectrum {
	for i := range rgb.c {
		rgb.c[i] /= sp.At(i)
	}

	return rgb
}

func (rgb *RGB) Pow(p float64) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] = math.Pow(rgb.c[i], p)
	}

	return &res
}

func (rgb *RGB) Exp() Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] = math.Exp(rgb.c[i])
	}

	return &res
}

func (rgb *RGB) Clamp(low, high float64) Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] = common.ClampF(rgb.c[i], low, high)
	}

	return &res
}

func (rgb *RGB) Sqrt() Spectrum {
	res := *rgb
	for i := range rgb.c {
		res.c[i] = math.Sqrt(rgb.c[i])
	}

	return &res
}

func (rgb *RGB) IsBlack() bool {
	for i := range rgb.c {
		if rgb.c[i] != 0.0 {
			return false
		}
	}

	return true
}

func (rgb *RGB) HasNaNs() bool {
	for i := range rgb.c {
		if math.IsNaN(rgb.c[i]) {
			return true
		}
	}

	return false
}

func (rgb *RGB) Equal(sp Spectrum) bool {
	for i := range rgb.c {
		if rgb.c[i] != sp.At(i) {
			return false
		}
	}

	return true
}
