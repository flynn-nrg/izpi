package sampler

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*Albedo)(nil)

// Albedo represents an albedo sampler.
type Albedo struct{}

// NewAlbedo returns an instance of the albedo sampler.
func NewAlbedo() *Albedo {
	return &Albedo{}
}

func (a *Albedo) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		return mat.Albedo(rec.U(), rec.V(), rec.P())
	}
	return &vec3.Vec3Impl{}
}
