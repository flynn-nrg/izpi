package sampler

import (
	"math"
	"sync/atomic"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*Normal)(nil)

type Normal struct {
	numRays *uint64
}

func NewNormal(numRays *uint64) *Normal {
	return &Normal{
		numRays: numRays,
	}
}

func (n *Normal) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	atomic.AddUint64(n.numRays, 1)
	if rec, _, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		return rec.Normal()
	}
	return &vec3.Vec3Impl{}
}
