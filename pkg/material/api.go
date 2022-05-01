// Package material implements the different materials and their properties.
package material

import (
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/scatterrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Material defines the methods to handle materials.
type Material interface {
	Scatter(r ray.Ray, hr *hitrecord.HitRecord) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool)
	ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64
	IsEmitter() bool
	Emitted(rIn ray.Ray, rec *hitrecord.HitRecord, u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl
}
