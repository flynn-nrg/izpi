package sampler

import (
	"sync/atomic"

	"github.com/flynn-nrg/go-vfx/math32"
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/ray"
)

// Ensure interface compliance.
var _ Sampler = (*Normal)(nil)

type Normal struct {
	NonSpectral // Embed to get SampleSpectral method
	numRays     *uint64
}

func NewNormal(numRays *uint64) *Normal {
	return &Normal{
		NonSpectral: *NewNonSpectral(), // Initialize embedded struct
		numRays:     numRays,
	}
}

func (n *Normal) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.XorShift) *vec3.Vec3Impl {
	atomic.AddUint64(n.numRays, 1)
	if rec, _, ok := world.Hit(r, 0.001, math32.MaxFloat32); ok {
		return rec.Normal()
	}
	return &vec3.Vec3Impl{}
}
