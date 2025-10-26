package hitable

import (
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/segment"
)

// Ensure interface compliance.
var _ Hitable = (*XYRect)(nil)

// XYRect represents an axis aligned rectangle.
type XYRect struct {
	x0       float32
	x1       float32
	y0       float32
	y1       float32
	k        float32
	material material.Material
}

// NewXYRect returns an instance of an axis aligned rectangle.
func NewXYRect(x0 float32, x1 float32, y0 float32, y1 float32, k float32, mat material.Material) *XYRect {
	return &XYRect{
		x0:       x0,
		y0:       y0,
		x1:       x1,
		y1:       y1,
		k:        k,
		material: mat,
	}
}

func (xyr *XYRect) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	t := (xyr.k - r.Origin().Z) / r.Direction().Z
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	x := r.Origin().X + (t * r.Direction().X)
	y := r.Origin().Y + (t * r.Direction().Y)
	if x < xyr.x0 || x > xyr.x1 || y < xyr.y0 || y > xyr.y1 {
		return nil, nil, false
	}

	u := (x - xyr.x0) / (xyr.x1 - xyr.x0)
	v := (y - xyr.y0) / (xyr.y1 - xyr.y0)
	return hitrecord.New(t, u, v, r.PointAtParameter(t), vec3.Vec3Impl{Z: 1}), xyr.material, true
}

func (xyr *XYRect) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	rec, _, ok := xyr.Hit(r, tMin, tMax)
	if !ok {
		return nil, false, false
	}

	segments := []*segment.Segment{
		{
			A: vec3.Vec3Impl{X: xyr.x0, Y: xyr.y0, Z: xyr.k},
			B: vec3.Vec3Impl{X: xyr.x1, Y: xyr.y0, Z: xyr.k},
		},
		{
			A: vec3.Vec3Impl{X: xyr.x1, Y: xyr.y0, Z: xyr.k},
			B: vec3.Vec3Impl{X: xyr.x1, Y: xyr.y1, Z: xyr.k},
		},
		{
			A: vec3.Vec3Impl{X: xyr.x1, Y: xyr.y1, Z: xyr.k},
			B: vec3.Vec3Impl{X: xyr.x0, Y: xyr.y1, Z: xyr.k},
		},
		{
			A: vec3.Vec3Impl{X: xyr.x0, Y: xyr.y1, Z: xyr.k},
			B: vec3.Vec3Impl{X: xyr.x0, Y: xyr.y0, Z: xyr.k},
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

func (xyr *XYRect) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return aabb.New(
		vec3.Vec3Impl{
			X: xyr.x0,
			Y: xyr.y0,
			Z: xyr.k - 0.0001,
		},
		vec3.Vec3Impl{
			X: xyr.x1,
			Y: xyr.y1,
			Z: xyr.k + 0.001,
		}), true
}

func (xyr *XYRect) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float32 {
	return 0.0
}

func (xyr *XYRect) Random(o vec3.Vec3Impl, _ *fastrandom.XorShift) vec3.Vec3Impl {
	return vec3.Vec3Impl{X: 1}
}

func (xyr *XYRect) IsEmitter() bool {
	return xyr.material.IsEmitter()
}
