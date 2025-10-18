package material

import (
	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*DiffuseLight)(nil)

// DiffuseLight represents a diffuse light material.
type DiffuseLight struct {
	nonPBR
	nonPathLength
	nonWorldSetter
	emit         texture.Texture
	spectralEmit texture.SpectralTexture
}

// NewDiffuseLight returns an instance of a diffuse light.
func NewDiffuseLight(emit texture.Texture) *DiffuseLight {
	return &DiffuseLight{
		emit: emit,
	}
}

// NewSpectralDiffuseLight returns an instance of a diffuse light with spectral emission.
func NewSpectralDiffuseLight(spectralEmit texture.SpectralTexture) *DiffuseLight {
	return &DiffuseLight{
		spectralEmit: spectralEmit,
	}
}

// Scatter returns false for diffuse light materials.
func (dl *DiffuseLight) Scatter(_ ray.Ray, _ *hitrecord.HitRecord, _ *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	return nil, nil, false
}

// SpectralScatter returns false for diffuse light materials.
func (dl *DiffuseLight) SpectralScatter(_ ray.Ray, _ *hitrecord.HitRecord, _ *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	return nil, nil, false
}

// Emitted returns the texture value at that point.
func (dl *DiffuseLight) Emitted(rIn ray.Ray, rec *hitrecord.HitRecord, u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	if vec3.Dot(rec.Normal(), rIn.Direction()) < 0.0 {
		return dl.emit.Value(u, v, p)
	}

	return &vec3.Vec3Impl{}
}

// EmittedSpectral returns the spectral emission at the given wavelength for diffuse lights.
func (dl *DiffuseLight) EmittedSpectral(rIn ray.Ray, rec *hitrecord.HitRecord, u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32 {
	if vec3.Dot(rec.Normal(), rIn.Direction()) < 0.0 {
		return dl.spectralEmit.Value(u, v, lambda, p)
	}
	return 0.0
}

// ScatteringPDF implements the probability distribution function for diffuse lights.
func (dl *DiffuseLight) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float32 {
	return 0
}

func (dl *DiffuseLight) IsEmitter() bool {
	return true
}

func (dl *DiffuseLight) Albedo(u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return dl.emit.Value(u, v, p)
}

// SpectralAlbedo returns the spectral albedo at the given wavelength.
func (dl *DiffuseLight) SpectralAlbedo(u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32 {
	return dl.emit.Value(u, v, p).X // Use red component as approximation
}
