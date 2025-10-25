package material

import (
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

type nonEmitter struct{}

// Emitted returns black for non-emitter materials.
func (ne *nonEmitter) Emitted(_ ray.Ray, _ *hitrecord.HitRecord, _ float64, _ float64, _ vec3.Vec3Impl) vec3.Vec3Impl {
	return vec3.Vec3Impl{}
}

func (ne *nonEmitter) IsEmitter() bool {
	return false
}

// EmittedSpectral returns 0.0 for non-emitter materials.
func (ne *nonEmitter) EmittedSpectral(_ ray.Ray, _ *hitrecord.HitRecord, _ float64, _ float64, _ float64, _ vec3.Vec3Impl) float64 {
	return 0.0
}
