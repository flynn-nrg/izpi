package sampler

import (
	"math"
	"sync/atomic"

	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*WireFrame)(nil)

// WireFrame represents a WireFrame sampler.
type WireFrame struct {
	NonSpectral // Embed to get SampleSpectral method
	numRays     *uint64
	paper       *vec3.Vec3Impl
	ink         *vec3.Vec3Impl
}

// NewWireFrame returns a new wireframe sampler with the provided colour.
func NewWireFrame(paper, ink *vec3.Vec3Impl, numRays *uint64) *WireFrame {
	return &WireFrame{
		NonSpectral: *NewNonSpectral(), // Initialize embedded struct
		numRays:     numRays,
		paper:       paper,
		ink:         ink,
	}
}

func (w *WireFrame) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) *vec3.Vec3Impl {
	atomic.AddUint64(w.numRays, 1)
	if _, _, ok := world.HitEdge(r, 0.001, math.Maxfloat32); ok {
		return w.ink
	}
	return w.paper
}
