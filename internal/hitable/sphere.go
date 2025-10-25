package hitable

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/onb"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Sphere)(nil)

// Sphere represents a sphere in the 3d world.
type Sphere struct {
	center0  vec3.Vec3Impl
	center1  vec3.Vec3Impl
	time0    float64
	time1    float64
	radius   float64
	material material.Material
}

func getSphereUV(p vec3.Vec3Impl) (float64, float64) {
	phi := math.Atan2(p.Z, p.X)
	theta := math.Asin(p.Y)
	u := 1.0 - (phi+math.Pi)/(2.0*math.Pi)
	v := (theta + math.Pi/2.0) / math.Pi
	return u, v
}

// SkyDome is a convenience function to construct a light emitting sphere with inverted normals.
// For this to work correctly texture file needs to be in HDR (Radiance) format.
func NewSkyDome(center vec3.Vec3Impl, radius float64, fileName string) (*FlipNormals, error) {
	texture, err := texture.NewFromHDR(fileName)
	texture.FlipY()
	texture.FlipX()
	if err != nil {
		return nil, err
	}
	light := material.NewDiffuseLight(texture)
	return NewFlipNormals(NewSphere(center, center, 0, 1, radius, light)), nil
}

// NewSphere returns a new instance of Sphere.
func NewSphere(center0 vec3.Vec3Impl, center1 vec3.Vec3Impl, time0 float64, time1 float64, radius float64, material material.Material) *Sphere {
	return &Sphere{
		center0:  center0,
		center1:  center1,
		time0:    time0,
		time1:    time1,
		radius:   radius,
		material: material,
	}
}

// Hit computes whether a ray intersects with the defined sphere.
func (s *Sphere) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	oc := vec3.Sub(r.Origin(), s.center(r.Time()))
	a := vec3.Dot(r.Direction(), r.Direction())
	b := vec3.Dot(oc, r.Direction())
	c := vec3.Dot(oc, oc) - (s.radius * s.radius)

	discriminant := (b * b) - (a * c)
	if discriminant > 0 {
		temp := (-b - math.Sqrt(b*b-a*c)) / a
		if temp < tMax && temp > tMin {
			outwardNormal := vec3.ScalarDiv(vec3.Sub(r.PointAtParameter(temp), s.center(r.Time())), s.radius)
			if vec3.Dot(r.Direction(), outwardNormal) >= 0 {
				outwardNormal = vec3.ScalarMul(outwardNormal, -1)
			}
			u, v := getSphereUV(outwardNormal)
			return hitrecord.New(temp, u, v, r.PointAtParameter(temp),
				outwardNormal), s.material, true
		}

		temp = (-b + math.Sqrt(b*b-a*c)) / a
		if temp < tMax && temp > tMin {
			outwardNormal := vec3.ScalarDiv(vec3.Sub(r.PointAtParameter(temp), s.center(r.Time())), s.radius)
			if vec3.Dot(r.Direction(), outwardNormal) >= 0 {
				outwardNormal = vec3.ScalarMul(outwardNormal, -1)
			}
			u, v := getSphereUV(outwardNormal)
			return hitrecord.New(temp, u, v,
				r.PointAtParameter(temp),
				vec3.ScalarDiv(vec3.Sub(r.PointAtParameter(temp), s.center(r.Time())), s.radius)), s.material, true
		}
	}
	return nil, nil, false
}

func (s *Sphere) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
	rec, _, ok := s.Hit(r, tMin, tMax)
	if !ok {
		return nil, false, false
	}

	a := vec3.Sub(rec.P(), r.Origin())
	b := vec3.Sub(rec.P(), s.center(r.Time()))

	ab := vec3.Dot(a, b)
	theta := math.Acos(ab / (a.Length() * b.Length()))
	if math.Abs(theta) <= (math.Pi/2.0 + 0.1) {
		return rec, true, true
	}

	return rec, true, false
}

func (s *Sphere) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	box0 := aabb.New(
		vec3.Sub(s.center0, vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}),
		vec3.Add(s.center0, vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}))
	box1 := aabb.New(
		vec3.Sub(s.center1, vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}),
		vec3.Add(s.center1, vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}))
	return aabb.SurroundingBox(box0, box1), true
}

func (s *Sphere) center(time float64) vec3.Vec3Impl {
	return vec3.Add(s.center0, vec3.ScalarMul(vec3.Sub(s.center1, s.center0), ((time-s.time0)/(s.time1-s.time0))))
}

func (s *Sphere) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float64 {
	if _, _, ok := s.Hit((ray.New(o, v, 0)), 0.001, math.MaxFloat64); ok {
		cosThetaMax := math.Sqrt(1 - s.radius*s.radius/vec3.Sub(s.center0, o).SquaredLength())
		solidAngle := 2 * math.Pi * (1 - cosThetaMax)
		return 1 / solidAngle
	}

	return 0.0
}

func (s *Sphere) Random(o vec3.Vec3Impl, random *fastrandom.LCG) vec3.Vec3Impl {
	direction := vec3.Sub(s.center0, o)
	distanceSquared := direction.SquaredLength()
	uvw := onb.New()
	uvw.BuildFromW(direction)
	return uvw.Local(vec3.RandomToSphere(s.radius, distanceSquared, random))
}

func (s *Sphere) IsEmitter() bool {
	return s.material.IsEmitter()
}
