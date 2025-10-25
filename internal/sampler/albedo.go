package sampler

import (
	"math"
	"sync/atomic"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*Albedo)(nil)

// Albedo represents an albedo sampler.
type Albedo struct {
	NonSpectral // Embed to get SampleSpectral method
	numRays     *uint64
}

// NewAlbedo returns an instance of the albedo sampler.
func NewAlbedo(numRays *uint64) *Albedo {
	return &Albedo{
		NonSpectral: *NewNonSpectral(), // Initialize embedded struct
		numRays:     numRays,
	}
}

func (a *Albedo) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) vec3.Vec3Impl {
	atomic.AddUint64(a.numRays, 1)
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		return mat.Albedo(rec.U(), rec.V(), rec.P())
	}
	return vec3.Vec3Impl{}
}
