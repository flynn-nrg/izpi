// Package material implements the different materials and their properties.
package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Material defines the methods to handle materials.
type Material interface {
	Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool)
	NormalMap() texture.Texture
	Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl
	ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64
	IsEmitter() bool
	Emitted(rIn ray.Ray, rec *hitrecord.HitRecord, u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl
}

// SpectralMaterial defines the methods to handle spectral materials.
type SpectralMaterial interface {
	Albedo(u float64, v float64, lambda float64, p *vec3.Vec3Impl) float64
	Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool)
	ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64
	IsEmitter() bool
	NormalMap() texture.Texture
}
