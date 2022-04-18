package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*YZRect)(nil)

// YZRect represents an axis aligned rectangle.
type YZRect struct {
	y0       float64
	y1       float64
	z0       float64
	z1       float64
	k        float64
	material material.Material
}

// NewYZRect returns an instance of an axis aligned rectangle.
func NewYZRect(y0 float64, y1 float64, z0 float64, z1 float64, k float64, mat material.Material) *YZRect {
	return &YZRect{
		y0:       y0,
		z0:       z0,
		y1:       y1,
		z1:       z1,
		k:        k,
		material: mat,
	}
}

func (xyr *YZRect) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	t := (xyr.k - r.Origin().X) / r.Direction().X
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	y := r.Origin().Y + (t * r.Direction().Y)
	z := r.Origin().Z + (t * r.Direction().Z)
	if y < xyr.y0 || y > xyr.y1 || z < xyr.z0 || z > xyr.z1 {
		return nil, nil, false
	}

	u := (y - xyr.y0) / (xyr.y1 - xyr.y0)
	v := (z - xyr.z0) / (xyr.z1 - xyr.z0)
	return hitrecord.New(t, u, v, r.PointAtParameter(t), &vec3.Vec3Impl{X: 1}), xyr.material, true
}

func (xyr *YZRect) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return aabb.New(
		&vec3.Vec3Impl{
			X: xyr.k - 0.0001,
			Y: xyr.y0,
			Z: xyr.z0,
		},
		&vec3.Vec3Impl{
			X: xyr.k + 0.001,
			Y: xyr.y1,
			Z: xyr.z1,
		}), true
}
