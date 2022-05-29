package sampler

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*Normal)(nil)

type Normal struct{}

func NewNormal() *Normal {
	return &Normal{}
}

func (n *Normal) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if rec, _, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		normal := rec.Normal()
		normal.X = math.Abs(normal.X)
		normal.Y = math.Abs(normal.Y)
		normal.Z = math.Abs(normal.Z)
		return normal
	}
	return &vec3.Vec3Impl{}
}
