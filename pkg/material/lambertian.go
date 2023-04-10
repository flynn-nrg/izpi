package material

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/fastrandom"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/onb"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/scatterrecord"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*Lambertian)(nil)

// Lambertian represents a diffuse material.
type Lambertian struct {
	nonPBR
	nonEmitter
	albedo texture.Texture
}

// NewLambertian returns an instance of the Lambert material.
func NewLambertian(albedo texture.Texture) *Lambertian {
	return &Lambertian{
		albedo: albedo,
	}
}

// Scatter computes how the ray bounces off the surface of a diffuse material.
func (l *Lambertian) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	uvw := onb.New()
	uvw.BuildFromW(hr.Normal())
	direction := uvw.Local(vec3.RandomCosineDirection(random))
	scattered := ray.New(hr.P(), vec3.UnitVector(direction), r.Time())
	albedo := l.albedo.Value(hr.U(), hr.V(), hr.P())
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.New(nil, false, albedo, nil, nil, nil, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for diffuse materials.
func (l *Lambertian) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	cosine := vec3.Dot(hr.Normal(), vec3.UnitVector(scattered.Direction()))
	if cosine < 0 {
		cosine = 0
	}

	return cosine / math.Pi
}

func (l *Lambertian) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return l.albedo.Value(u, v, p)
}
