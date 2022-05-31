package material

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/onb"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/scatterrecord"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*PBR)(nil)

type PBR struct {
	nonEmitter
	albedo    texture.Texture
	normalMap texture.Texture
	roughness texture.Texture
	metalness texture.Texture
}

// NewPBR returns a new PBR material with the supplied textures.
func NewPBR(albedo, normalMap, roughness, metalness texture.Texture) *PBR {
	return &PBR{
		albedo:    albedo,
		normalMap: normalMap,
		roughness: roughness,
		metalness: metalness,
	}
}

// Scatter computes how the ray bounces off the surface of a PBR material.
func (pbr *PBR) Scatter(r ray.Ray, hr *hitrecord.HitRecord) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {

	uvw := onb.New()
	uvw.BuildFromW(hr.Normal())
	direction := uvw.Local(vec3.RandomCosineDirection())
	scattered := ray.New(hr.P(), vec3.UnitVector(direction), r.Time())
	albedo := pbr.albedo.Value(hr.U(), hr.V(), hr.P())
	normalAtUV := pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
	roughness := pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	metalness := pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.New(nil, false, albedo, normalAtUV, roughness, metalness, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for PBR materials.
func (pbr *PBR) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	cosine := vec3.Dot(hr.Normal(), vec3.UnitVector(scattered.Direction()))
	if cosine < 0 {
		cosine = 0
	}

	return cosine / math.Pi
}

// NormalMap() returns the normal map associated with this material.
func (pbr *PBR) NormalMap() texture.Texture {
	return pbr.normalMap
}

func (pbr *PBR) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return pbr.albedo.Value(u, v, p)
}
