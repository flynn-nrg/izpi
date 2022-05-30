package hitable

import (
	"math"
	"math/rand"

	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/segment"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*XZRect)(nil)

// XZRect represents an axis aligned rectangle.
type XZRect struct {
	x0       float64
	x1       float64
	z0       float64
	z1       float64
	k        float64
	material material.Material
}

// NewXZRect returns an instance of an axis aligned rectangle.
func NewXZRect(x0 float64, x1 float64, z0 float64, z1 float64, k float64, mat material.Material) *XZRect {
	return &XZRect{
		x0:       x0,
		z0:       z0,
		x1:       x1,
		z1:       z1,
		k:        k,
		material: mat,
	}
}

func (xzr *XZRect) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	t := (xzr.k - r.Origin().Y) / r.Direction().Y
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	x := r.Origin().X + (t * r.Direction().X)
	z := r.Origin().Z + (t * r.Direction().Z)
	if x < xzr.x0 || x > xzr.x1 || z < xzr.z0 || z > xzr.z1 {
		return nil, nil, false
	}

	u := (x - xzr.x0) / (xzr.x1 - xzr.x0)
	v := (z - xzr.z0) / (xzr.z1 - xzr.z0)
	return hitrecord.New(t, u, v, r.PointAtParameter(t), &vec3.Vec3Impl{Y: 1}), xzr.material, true
}

func (xzr *XZRect) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
	rec, _, ok := xzr.Hit(r, tMin, tMax)
	if !ok {
		return nil, false, false
	}

	segments := []*segment.Segment{
		{
			A: &vec3.Vec3Impl{X: xzr.x0, Y: xzr.k, Z: xzr.z0},
			B: &vec3.Vec3Impl{X: xzr.x1, Y: xzr.k, Z: xzr.z0},
		},
		{
			A: &vec3.Vec3Impl{X: xzr.x1, Y: xzr.k, Z: xzr.z0},
			B: &vec3.Vec3Impl{X: xzr.x1, Y: xzr.k, Z: xzr.z1},
		},
		{
			A: &vec3.Vec3Impl{X: xzr.x1, Y: xzr.k, Z: xzr.z1},
			B: &vec3.Vec3Impl{X: xzr.x0, Y: xzr.k, Z: xzr.z1},
		},
		{
			A: &vec3.Vec3Impl{X: xzr.x0, Y: xzr.k, Z: xzr.z1},
			B: &vec3.Vec3Impl{X: xzr.x0, Y: xzr.k, Z: xzr.z0},
		},
	}

	c := rec.P()
	for _, s := range segments {
		if segment.Belongs(s, c) {
			return rec, true, true
		}
	}

	return nil, true, false
}

func (xzr *XZRect) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return aabb.New(
		&vec3.Vec3Impl{
			X: xzr.x0,
			Y: xzr.k - 0.0001,
			Z: xzr.z0,
		},
		&vec3.Vec3Impl{
			X: xzr.x1,
			Y: xzr.k + 0.001,
			Z: xzr.z1,
		}), true
}

func (xzr *XZRect) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	r := ray.New(o, v, 0)
	if rec, _, ok := xzr.Hit(r, 0.001, math.MaxFloat64); ok {
		area := (xzr.x1 - xzr.x0) * (xzr.z1 - xzr.z0)
		distanceSquared := rec.T() * rec.T() * v.SquaredLength()
		cosine := math.Abs(vec3.Dot(v, vec3.ScalarDiv(rec.Normal(), v.Length())))
		return distanceSquared / (cosine * area)
	}

	return 0
}

func (xzr *XZRect) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	randomPoint := &vec3.Vec3Impl{
		X: xzr.x0 + rand.Float64()*(xzr.x1-xzr.x0),
		Y: xzr.k,
		Z: xzr.z0 + rand.Float64()*(xzr.z1-xzr.z0),
	}

	return vec3.Sub(randomPoint, o)
}

func (xzr *XZRect) IsEmitter() bool {
	return xzr.material.IsEmitter()
}
