package material

import (
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/scatterrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/texture"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*DiffuseLight)(nil)

// DiffuseLight represents a diffuse light material.
type DiffuseLight struct {
	nonPBR
	emit texture.Texture
}

// NewDiffuseLight returns an instance of a diffuse light.
func NewDiffuseLight(emit texture.Texture) *DiffuseLight {
	return &DiffuseLight{
		emit: emit,
	}
}

// Scatter returns false for diffuse light materials.
func (dl *DiffuseLight) Scatter(_ ray.Ray, _ *hitrecord.HitRecord) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	return nil, nil, false
}

// Emitted returns the texture value at that point.
func (dl *DiffuseLight) Emitted(rIn ray.Ray, rec *hitrecord.HitRecord, u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	if vec3.Dot(rec.Normal(), rIn.Direction()) < 0.0 {
		return dl.emit.Value(u, v, p)
	}

	return &vec3.Vec3Impl{}
}

// ScatteringPDF implements the probability distribution function for diffuse lights.
func (dl *DiffuseLight) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	return 0
}

func (dl *DiffuseLight) IsEmitter() bool {
	return true
}
