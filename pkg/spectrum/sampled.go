package spectrum

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/common"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

const (
	sampledLambdaStart = 400
	sampledLambdaEnd   = 700
	nSpectralSamples   = 60
)

// Ensure interface compliance.
var _ Spectrum = (*Sampled)(nil)

type Sampled struct {
	c                     []float64
	x                     *Sampled
	y                     *Sampled
	z                     *Sampled
	rgbRefl2SpectWhite    *Sampled
	rgbRefl2SpectCyan     *Sampled
	rgbRefl2SpectMagenta  *Sampled
	rgbRefl2SpectYellow   *Sampled
	rgbRefl2SpectRed      *Sampled
	rgbRefl2SpectGreen    *Sampled
	rgbRefl2SpectBlue     *Sampled
	rgbIllum2SpectWhite   *Sampled
	rgbIllum2SpectCyan    *Sampled
	rgbIllum2SpectMagenta *Sampled
	rgbIllum2SpectYellow  *Sampled
	rgbIllum2SpectRed     *Sampled
	rgbIllum2SpectGreen   *Sampled
	rgbIllum2SpectBlue    *Sampled
}

// SampledFromSampled returns a new Sampled spectrum. The lambdas
// and values slices must be sorted.
func SampledFromSampled(lambdas, values []float64) *Sampled {
	c := make([]float64, nSpectralSamples)
	for i := 0; i < nSpectralSamples; i++ {
		lambda0 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i)/nSpectralSamples)
		lambda1 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i+1)/nSpectralSamples)
		c[i] = averageSpectrumSamples(lambdas, values, len(lambdas), lambda0, lambda1)
	}

	sampled := &Sampled{}
	sampled.c = c
	sampled.initXYZ()
	sampled.initRGB()
	return sampled
}

func SampledFromRGB(rgb *vec3.Vec3Impl, spectrumType SpectrumType) *Sampled {
	sampled := &Sampled{}
	sampled.initXYZ()
	sampled.initRGB()
	sampled.c = make([]float64, nSpectralSamples)

	switch spectrumType {
	case Reflectance:
		if rgb.X <= rgb.Y && rgb.X <= rgb.Z {
			sampled = sampled.Add(sampled.rgbRefl2SpectWhite.ScalarMul(rgb.X)).(*Sampled)
			if rgb.Y <= rgb.Z {
				sampled = sampled.Add(sampled.rgbRefl2SpectCyan.ScalarMul(rgb.Y - rgb.X)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectBlue.ScalarMul(rgb.Z - rgb.Y)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbRefl2SpectCyan.ScalarMul(rgb.Z - rgb.X)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectGreen.ScalarMul(rgb.Y - rgb.Z)).(*Sampled)
			}
		} else if rgb.Y <= rgb.X && rgb.Y <= rgb.Z {
			sampled = sampled.Add(sampled.rgbRefl2SpectWhite.ScalarMul(rgb.Y)).(*Sampled)
			if rgb.X <= rgb.Z {
				sampled = sampled.Add(sampled.rgbRefl2SpectMagenta.ScalarMul(rgb.X - rgb.Y)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectBlue.ScalarMul(rgb.Z - rgb.X)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbRefl2SpectMagenta.ScalarMul(rgb.Z - rgb.Y)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectRed.ScalarMul(rgb.X - rgb.Z)).(*Sampled)
			}
		} else {
			sampled = sampled.Add(sampled.rgbRefl2SpectWhite.ScalarMul(rgb.Z)).(*Sampled)
			if rgb.X <= rgb.Y {
				sampled = sampled.Add(sampled.rgbRefl2SpectYellow.ScalarMul(rgb.X - rgb.Z)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectGreen.ScalarMul(rgb.Y - rgb.X)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbRefl2SpectYellow.ScalarMul(rgb.Y - rgb.Z)).(*Sampled)
				sampled = sampled.Add(sampled.rgbRefl2SpectRed.ScalarMul(rgb.X - rgb.Y)).(*Sampled)
			}
		}
	case Illuminant:
		if rgb.X <= rgb.Y && rgb.X <= rgb.Z {
			sampled = sampled.Add(sampled.rgbIllum2SpectWhite.ScalarMul(rgb.X)).(*Sampled)
			if rgb.Y <= rgb.Z {
				sampled = sampled.Add(sampled.rgbIllum2SpectCyan.ScalarMul(rgb.Y - rgb.X)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectBlue.ScalarMul(rgb.Z - rgb.Y)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbIllum2SpectCyan.ScalarMul(rgb.Z - rgb.X)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectGreen.ScalarMul(rgb.Y - rgb.Z)).(*Sampled)
			}
		} else if rgb.Y <= rgb.X && rgb.Y <= rgb.Z {
			sampled = sampled.Add(sampled.rgbIllum2SpectWhite.ScalarMul(rgb.Y)).(*Sampled)
			if rgb.X <= rgb.Z {
				sampled = sampled.Add(sampled.rgbIllum2SpectMagenta.ScalarMul(rgb.X - rgb.Y)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectBlue.ScalarMul(rgb.Z - rgb.X)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbIllum2SpectMagenta.ScalarMul(rgb.Z - rgb.Y)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectRed.ScalarMul(rgb.X - rgb.Z)).(*Sampled)
			}
		} else {
			sampled = sampled.Add(sampled.rgbIllum2SpectWhite.ScalarMul(rgb.Z)).(*Sampled)
			if rgb.X <= rgb.Y {
				sampled = sampled.Add(sampled.rgbIllum2SpectYellow.ScalarMul(rgb.X - rgb.Z)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectGreen.ScalarMul(rgb.Y - rgb.X)).(*Sampled)
			} else {
				sampled = sampled.Add(sampled.rgbIllum2SpectYellow.ScalarMul(rgb.Y - rgb.Z)).(*Sampled)
				sampled = sampled.Add(sampled.rgbIllum2SpectRed.ScalarMul(rgb.X - rgb.Y)).(*Sampled)
			}
		}
	}

	sampled = sampled.Clamp(0, math.MaxFloat64).(*Sampled)
	return sampled
}

func averageSpectrumSamples(lambdas, values []float64, n int, lambdaStart float64, lambdaEnd float64) float64 {
	if lambdaEnd <= lambdas[0] {
		return values[0]
	}

	if lambdaStart >= lambdas[n-1] {
		return values[n-1]
	}

	if n == 1 {
		return values[0]
	}

	sum := 0.0

	if lambdaStart < lambdas[0] {
		sum += values[0] * (lambdas[0] - lambdaStart)
	}

	if lambdaEnd > lambdas[n-1] {
		sum += values[n-1] * (lambdaEnd - lambdas[n-1])
	}

	i := 1
	for {
		if lambdaStart <= lambdas[i] {
			break
		}
		i++
	}

	for ; i+1 < n && lambdaEnd >= lambdas[i]; i++ {
		segLambdaStart := common.MaxF(lambdaStart, float64(lambdas[i]))
		segLambdaEnd := common.MinF(lambdaEnd, float64(lambdas[i+1]))

		a := common.Lerp(values[i], values[i+1], (segLambdaStart - float64(lambdas[i])/(lambdas[i+1]-lambdas[i])))
		b := common.Lerp(values[i], values[i+1], (segLambdaEnd - float64(lambdas[i])/(lambdas[i+1]-lambdas[i])))
		sum += (0.5 * a) + b*(segLambdaEnd-segLambdaStart)
	}

	return sum / (lambdaEnd - lambdaStart)
}

func (ss *Sampled) ToXYZ() *vec3.Vec3Impl {
	v := &vec3.Vec3Impl{}
	for i := range ss.c {
		v.X += ss.x.c[i] * ss.c[i]
		v.Y += ss.y.c[i] * ss.c[i]
		v.Z += ss.z.c[i] * ss.c[i]

	}

	scale := 1.0 // float64(sampledLambdaEnd - sampledLambdaStart) / float64(CIE_Y_integral * nSpectralSamples)

	v.X *= scale
	v.Y *= scale
	v.Z *= scale

	return v
}

func (ss *Sampled) Y() float64 {
	var yy float64
	for i := range ss.y.c {
		yy += ss.y.c[i] * ss.c[i]
	}

	return yy * float64(sampledLambdaEnd-sampledLambdaStart) / float64(nSpectralSamples)
}

func (ss *Sampled) ToRGB() *vec3.Vec3Impl {
	xyz := ss.ToXYZ()
	return XYZToRGB(xyz)
}

func (ss *Sampled) ToRGBSpectrum() *RGB {
	rgb := ss.ToRGB()
	return RGBFromRGB(rgb, Reflectance)
}

func (ss *Sampled) initRGB() {
	ss.rgbRefl2SpectWhite.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectCyan.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectMagenta.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectYellow.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectRed.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectGreen.c = make([]float64, nSpectralSamples)
	ss.rgbRefl2SpectBlue.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectWhite.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectCyan.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectMagenta.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectYellow.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectRed.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectGreen.c = make([]float64, nSpectralSamples)
	ss.rgbIllum2SpectBlue.c = make([]float64, nSpectralSamples)
	for i := 0; i < nSpectralSamples; i++ {
		wl0 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i)/float64(nSpectralSamples))
		wl1 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i+1)/float64(nSpectralSamples))
		ss.rgbRefl2SpectWhite.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectWhite,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectCyan.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectCyan,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectMagenta.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectMagenta,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectYellow.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectYellow,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectRed.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectRed,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectGreen.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectGreen,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbRefl2SpectBlue.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbRefl2SpectBlue,
			nRGB2SpectSamples, wl0, wl1)

		ss.rgbIllum2SpectWhite.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectWhite,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectCyan.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectCyan,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectMagenta.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectMagenta,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectYellow.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectYellow,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectRed.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectRed,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectGreen.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectGreen,
			nRGB2SpectSamples, wl0, wl1)
		ss.rgbIllum2SpectBlue.c[i] = averageSpectrumSamples(rgb2SpectLambda, rgbIllum2SpectBlue,
			nRGB2SpectSamples, wl0, wl1)
	}
}

func (ss *Sampled) initXYZ() {
	for i := 0; i < nSpectralSamples; i++ {
		wl0 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i)/float64(nSpectralSamples))
		wl1 := common.Lerp(sampledLambdaStart, sampledLambdaEnd, float64(i+1)/float64(nSpectralSamples))
		ss.x.c[i] = averageSpectrumSamples(cieLambda, cieX, nCIESamples, wl0, wl1)
		ss.y.c[i] = averageSpectrumSamples(cieLambda, cieY, nCIESamples, wl0, wl1)
		ss.z.c[i] = averageSpectrumSamples(cieLambda, cieZ, nCIESamples, wl0, wl1)
	}
}

func (ss *Sampled) At(i int) float64 {
	return ss.c[i]
}

func (ss *Sampled) Add(sp Spectrum) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] += sp.At(i)
	}

	return &res
}

func (ss *Sampled) Sub(sp Spectrum) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] -= sp.At(i)
	}

	return &res
}

func (ss *Sampled) Mul(sp Spectrum) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] *= sp.At(i)
	}

	return &res
}

func (ss *Sampled) ScalarMul(v float64) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] *= v
	}

	return &res
}

func (ss *Sampled) Div(sp Spectrum) Spectrum {
	for i := range ss.c {
		ss.c[i] /= sp.At(i)
	}

	return ss
}

func (ss *Sampled) Pow(p float64) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] = math.Pow(ss.c[i], p)
	}

	return &res
}

func (ss *Sampled) Exp() Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] = math.Exp(ss.c[i])
	}

	return &res
}

func (ss *Sampled) Clamp(low, high float64) Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] = common.ClampF(ss.c[i], low, high)
	}

	return &res
}

func (ss *Sampled) Sqrt() Spectrum {
	res := *ss
	for i := range ss.c {
		res.c[i] = math.Sqrt(ss.c[i])
	}

	return &res
}

func (ss *Sampled) IsBlack() bool {
	for i := range ss.c {
		if ss.c[i] != 0.0 {
			return false
		}
	}

	return true
}

func (ss *Sampled) HasNaNs() bool {
	for i := range ss.c {
		if math.IsNaN(ss.c[i]) {
			return true
		}
	}

	return false
}

func (ss *Sampled) Equal(sp Spectrum) bool {
	for i := range ss.c {
		if ss.c[i] != sp.At(i) {
			return false
		}
	}

	return true
}
