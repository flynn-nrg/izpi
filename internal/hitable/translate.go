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
var _ Hitable = (*Translate)(nil)

// Translate represents a hitable with its associated translation.
type Translate struct {
	hitable Hitable
	offset  vec3.Vec3Impl
}

// NewTranslate returns an instance of a translated hitable.
func NewTranslate(hitable Hitable, offset vec3.Vec3Impl) *Translate {
	return &Translate{
		hitable: hitable,
		offset:  offset,
	}
}

func (tr *Translate) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	movedRay := ray.New(vec3.Sub(r.Origin(), tr.offset), r.Direction(), r.Time())
	if hr, mat, ok := tr.hitable.Hit(movedRay, tMin, tMax); ok {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), vec3.Add(hr.P(), tr.offset), hr.Normal()), mat, true
	}

	return nil, nil, false
}

func (tr *Translate) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	movedRay := ray.New(vec3.Sub(r.Origin(), tr.offset), r.Direction(), r.Time())
	hr, hitOk, edgeOk := tr.hitable.HitEdge(movedRay, tMin, tMax)
	if hitOk {
		return hitrecord.New(hr.T(), hr.U(), hr.V(), vec3.Add(hr.P(), tr.offset), hr.Normal()), true, edgeOk
	}

	return nil, false, false
}

func (tr *Translate) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	if bbox, ok := tr.hitable.BoundingBox(time0, time1); ok {
		return aabb.New(vec3.Add(bbox.Min(), tr.offset), vec3.Add(bbox.Max(), tr.offset)), true
	}

	return nil, false
}

func (tr *Translate) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float32 {
	return tr.hitable.PDFValue(o, v)
}

func (tr *Translate) Random(o vec3.Vec3Impl, random *fastrandom.XorShift) vec3.Vec3Impl {
	return tr.hitable.Random(o, random)
}

func (tr *Translate) IsEmitter() bool {
	return tr.hitable.IsEmitter()
}
