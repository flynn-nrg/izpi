package hitable

import (
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/segment"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*YZRect)(nil)

// YZRect represents an axis aligned rectangle.
type YZRect struct {
	y0       float32
	y1       float32
	z0       float32
	z1       float32
	k        float32
	material material.Material
}

// NewYZRect returns an instance of an axis aligned rectangle.
func NewYZRect(y0 float32, y1 float32, z0 float32, z1 float32, k float32, mat material.Material) *YZRect {
	return &YZRect{
		y0:       y0,
		z0:       z0,
		y1:       y1,
		z1:       z1,
		k:        k,
		material: mat,
	}
}

func (yzr *YZRect) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
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

func (yzr *YZRect) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	rec, _, ok := yzr.Hit(r, tMin, tMax)
	if !ok {
		return nil, false, false
	}

	segments := []*segment.Segment{
		{
			A: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y0, Z: yzr.z0},
			B: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y0, Z: yzr.z1},
		},
		{
			A: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y0, Z: yzr.z1},
			B: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y1, Z: yzr.z1},
		},
		{
			A: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y1, Z: yzr.z1},
			B: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y1, Z: yzr.z0},
		},
		{
			A: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y1, Z: yzr.z0},
			B: &vec3.Vec3Impl{X: yzr.k, Y: yzr.y0, Z: yzr.z0},
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

func (yzr *YZRect) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
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

func (yzr *YZRect) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	return 0.0
}

func (yzr *YZRect) Random(o *vec3.Vec3Impl, _ *fastrandom.LCG) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}

func (yzr *YZRect) IsEmitter() bool {
	return yzr.material.IsEmitter()
}
