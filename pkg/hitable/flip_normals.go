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
