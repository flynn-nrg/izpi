package hitable

import (
	"math"

	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Sphere)(nil)

// Sphere represents a sphere in the 3d world.
type Sphere struct {
	center0  *vec3.Vec3Impl
	center1  *vec3.Vec3Impl
	time0    float64
	time1    float64
	radius   float64
	material material.Material
}

// NewSphere returns a new instance of Sphere.
func NewSphere(center0 *vec3.Vec3Impl, center1 *vec3.Vec3Impl, time0 float64, time1 float64, radius float64, material material.Material) *Sphere {
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

func (s *Sphere) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	box0 := aabb.New(
		vec3.Sub(s.center0, &vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}),
		vec3.Add(s.center0, &vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}))
	box1 := aabb.New(
		vec3.Sub(s.center1, &vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}),
		vec3.Add(s.center1, &vec3.Vec3Impl{X: s.radius, Y: s.radius, Z: s.radius}))
	return aabb.SurroundingBox(box0, box1), true
}

func (s *Sphere) center(time float64) *vec3.Vec3Impl {
	return vec3.Add(s.center0, vec3.ScalarMul(vec3.Sub(s.center1, s.center0), ((time-s.time0)/(s.time1-s.time0))))
}

func getSphereUV(p *vec3.Vec3Impl) (float64, float64) {
	phi := math.Atan2(p.Z, p.X)
	theta := math.Asin(p.Y)
	u := 1.0 - (phi+math.Pi)/(2.0*math.Pi)
	v := (theta + math.Pi/2.0) / math.Pi
	return u, v
}
