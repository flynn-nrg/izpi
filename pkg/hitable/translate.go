package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Translate)(nil)

// Translate represents a hitable with its associated translation.
type Translate struct {
	hitable Hitable
	offset  *vec3.Vec3Impl
}

// NewTranslate returns an instance of a translated hitable.
func NewTranslate(hitable Hitable, offset *vec3.Vec3Impl) *Translate {
	return &Translate{
		hitable: hitable,
		offset:  offset,
	}
}

func (tr *Translate) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	movedRay := ray.New(vec3.Sub(r.Origin(), tr.offset), r.Direction(), r.Time())
	if hr, mat, ok := tr.hitable.Hit(movedRay, tMin, tMax); ok {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), vec3.Add(hr.P(), tr.offset), hr.Normal()), mat, true
	}

	return nil, nil, false
}

func (tr *Translate) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	if bbox, ok := tr.hitable.BoundingBox(time0, time1); ok {
		return aabb.New(vec3.Add(bbox.Min(), tr.offset), vec3.Add(bbox.Max(), tr.offset)), true
	}

	return nil, false
}
