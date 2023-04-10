package material

import (
	"github.com/flynn-nrg/izpi/pkg/fastrandom"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/scatterrecord"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*Isotropic)(nil)

// Isotropic represents an isotropic material.
type Isotropic struct {
	nonEmitter
	nonPBR
	albedo texture.Texture
}

// NewIsotropic returns a new instances of the isotropic material.
func NewIsotropic(albedo texture.Texture) *Isotropic {
	return &Isotropic{
		albedo: albedo,
	}
}

// Scatter computes how the ray bounces off the surface of a diffuse material.
func (i *Isotropic) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	scattered := ray.New(hr.P(), randomInUnitSphere(random), r.Time())
	attenuation := i.albedo.Value(hr.U(), hr.V(), hr.P())
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.New(nil, false, attenuation, nil, nil, nil, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for isotropic materials.
func (i *Isotropic) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	return 0
}

func (i *Isotropic) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return i.albedo.Value(u, v, p)
}
