package hitable

import (
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
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

func (fn *FlipNormals) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	if hr, mat, ok := fn.hitable.Hit(r, tMin, tMax); ok {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), hr.P(), vec3.ScalarMul(hr.Normal(), -1)), mat, true
	}
	return nil, nil, false
}

func (fn *FlipNormals) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	if hr, ok, edgeOk := fn.hitable.HitEdge(r, tMin, tMax); ok {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), hr.P(), vec3.ScalarMul(hr.Normal(), -1)), true, edgeOk
	}
	return nil, false, false
}

func (fn *FlipNormals) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return fn.hitable.BoundingBox(time0, time1)
}

func (fn *FlipNormals) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	return fn.hitable.PDFValue(o, v)
}

func (fn *FlipNormals) Random(o *vec3.Vec3Impl, random *fastrandom.XorShift) *vec3.Vec3Impl {
	return fn.hitable.Random(o, random)
}

func (fn *FlipNormals) IsEmitter() bool {
	return fn.hitable.IsEmitter()
}
