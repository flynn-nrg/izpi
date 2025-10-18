// Package sampler implements different types of samplers.
package sampler

import (
	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

type SamplerType int

const (
	InvalidSampler SamplerType = iota
	NormalSampler
	ColourSampler
	WireFrameSampler
	AlbedoSampler
	SpectralSampler
)

var samplerMap = map[string]SamplerType{
	"colour":    ColourSampler,
	"normal":    NormalSampler,
	"wireframe": WireFrameSampler,
	"albedo":    AlbedoSampler,
	"spectral":  SpectralSampler,
}

type Sampler interface {
	Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) *vec3.Vec3Impl
	SampleSpectral(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) float32
}

func StringToType(s string) SamplerType {
	return samplerMap[s]
}
