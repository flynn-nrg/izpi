package material

import (
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

type nonEmitter struct{}

// Emitted returns black for non-emitter materials.
func (ne *nonEmitter) Emitted(_ ray.Ray, _ *hitrecord.HitRecord, _ float64, _ float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{}
}

func (ne *nonEmitter) IsEmitter() bool {
	return false
}
