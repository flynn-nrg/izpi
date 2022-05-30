package sampler

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*WireFrame)(nil)

// WireFrame represents a WireFrame sampler.
type WireFrame struct {
	paper *vec3.Vec3Impl
	ink   *vec3.Vec3Impl
}

// NewWireFrame returns a new wireframe sampler with the provided colour.
func NewWireFrame(paper, ink *vec3.Vec3Impl) *WireFrame {
	return &WireFrame{
		paper: paper,
		ink:   ink,
	}
}

func (w *WireFrame) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if _, _, ok := world.HitEdge(r, 0.001, math.MaxFloat64); ok {
		return w.ink
	}
	return w.paper
}
