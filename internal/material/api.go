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

// SceneGeometry defines the minimal interface needed for path length calculation
type SceneGeometry interface {
	Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, Material, bool)
}

// Material defines the methods to handle materials.
type Material interface {
	Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool)
	SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool)
	NormalMap() texture.Texture
	Albedo(u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl
	SpectralAlbedo(u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32
	ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float32
	IsEmitter() bool
	Emitted(rIn ray.Ray, rec *hitrecord.HitRecord, u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl
	EmittedSpectral(rIn ray.Ray, rec *hitrecord.HitRecord, u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32
	SetWorld(world SceneGeometry)
}

// nonPathLength is a stub for materials that don't need path length calculation
type nonPathLength struct{}

func (npl *nonPathLength) CalculatePathLength(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray, world SceneGeometry) float32 {
	return 0.0
}

// nonWorldSetter is a stub for materials that don't need world reference
type nonWorldSetter struct{}

func (nws *nonWorldSetter) SetWorld(world SceneGeometry) {
	// Do nothing - this material doesn't need world reference
}
