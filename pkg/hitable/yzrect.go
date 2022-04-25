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

func (yzr *YZRect) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	t := (yzr.k - r.Origin().X) / r.Direction().X
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	y := r.Origin().Y + (t * r.Direction().Y)
	z := r.Origin().Z + (t * r.Direction().Z)
	if y < yzr.y0 || y > yzr.y1 || z < yzr.z0 || z > yzr.z1 {
		return nil, nil, false
	}

	u := (y - yzr.y0) / (yzr.y1 - yzr.y0)
	v := (z - yzr.z0) / (yzr.z1 - yzr.z0)
	return hitrecord.New(t, u, v, r.PointAtParameter(t), &vec3.Vec3Impl{X: 1}), yzr.material, true
}

func (yzr *YZRect) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return aabb.New(
		&vec3.Vec3Impl{
			X: yzr.k - 0.0001,
			Y: yzr.y0,
			Z: yzr.z0,
		},
		&vec3.Vec3Impl{
			X: yzr.k + 0.001,
			Y: yzr.y1,
			Z: yzr.z1,
		}), true
}

func (yzr *YZRect) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	return 0.0
}

func (yzr *YZRect) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}
