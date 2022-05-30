// Package sampler implements different types of samplers.
package sampler

import (
	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

type SamplerType int

const (
	InvalidSampler SamplerType = iota
	NormalSampler
	ColourSampler
	WireFrameSampler
)

var samplerMap = map[string]SamplerType{
	"colour":    ColourSampler,
	"normal":    NormalSampler,
	"wireframe": WireFrameSampler,
}

type Sampler interface {
	Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl
}

func StringToType(s string) SamplerType {
	return samplerMap[s]
}
