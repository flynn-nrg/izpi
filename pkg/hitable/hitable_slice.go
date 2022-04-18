package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
)

// Ensure interface compliance.
var _ Hitable = (*HitableSlice)(nil)

// HitableSlice represents a list of hitable entities.
type HitableSlice struct {
	hitables []Hitable
}

// NewSlice returns an instance of HitableSlice.
func NewSlice(hitables []Hitable) *HitableSlice {
	return &HitableSlice{
		hitables: hitables,
	}
}

// Hit computes whether a ray intersects with any of the elements in the slice.
func (hs *HitableSlice) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	var rec *hitrecord.HitRecord
	var mat material.Material
	var hitAnything bool
	closestSoFar := tMax

	for _, h := range hs.hitables {
		if tempRec, tempMat, ok := h.Hit(r, tMin, closestSoFar); ok {
			rec = tempRec
			mat = tempMat
			hitAnything = ok
			closestSoFar = rec.T()
		}
	}

	return rec, mat, hitAnything
}

func (hs *HitableSlice) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	var tempBox *aabb.AABB
	var box *aabb.AABB
	var ok bool

	if len(hs.hitables) < 1 {
		return nil, false
	}

	if tempBox, ok = hs.hitables[0].BoundingBox(time0, time1); ok {
		box = tempBox
	} else {
		return nil, false
	}

	for i := 1; i < len(hs.hitables); i++ {
		if tempBox, ok = hs.hitables[i].BoundingBox(time0, time1); ok {
			box = aabb.SurroundingBox(box, tempBox)
		} else {
			return nil, false
		}
	}

	return box, true
}
