package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*FlipNormals)(nil)

// FlipNormals represents a hitable with the inverted normal.
type FlipNormals struct {
	hitable Hitable
}

// NewFlipNormals returns a hitable with inverted normals.
func NewFlipNormals(hitable Hitable) *FlipNormals {
	return &FlipNormals{
		hitable: hitable,
	}
}

func (fn *FlipNormals) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	if hr, mat, ok := fn.hitable.Hit(r, tMin, tMax); ok {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), hr.P(), vec3.ScalarMul(hr.Normal(), -1)), mat, true
	}
	return nil, nil, false
}

func (fn *FlipNormals) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return fn.hitable.BoundingBox(time0, time1)
}

func (fn *FlipNormals) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	return fn.hitable.PDFValue(o, v)
}

func (fn *FlipNormals) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	return fn.hitable.Random(o)
}

func (fn *FlipNormals) IsEmitter() bool {
	return fn.hitable.IsEmitter()
}
