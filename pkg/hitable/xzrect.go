package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
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

func (xyr *XZRect) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	t := (xyr.k - r.Origin().Y) / r.Direction().Y
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	x := r.Origin().X + (t * r.Direction().X)
	z := r.Origin().Z + (t * r.Direction().Z)
	if x < xyr.x0 || x > xyr.x1 || z < xyr.z0 || z > xyr.z1 {
		return nil, nil, false
	}

	u := (x - xyr.x0) / (xyr.x1 - xyr.x0)
	v := (z - xyr.z0) / (xyr.z1 - xyr.z0)
	return hitrecord.New(t, u, v, r.PointAtParameter(t), &vec3.Vec3Impl{Y: 1}), xyr.material, true
}

func (xyr *XZRect) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return aabb.New(
		&vec3.Vec3Impl{
			X: xyr.x0,
			Y: xyr.k - 0.0001,
			Z: xyr.z0,
		},
		&vec3.Vec3Impl{
			X: xyr.x1,
			Y: xyr.k + 0.001,
			Z: xyr.z1,
		}), true
}
