package hitable

import (
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
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
func (hs *HitableSlice) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
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

// HitEdge computes whether a ray intersects with the edge of any of the elements in the slice.
func (hs *HitableSlice) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	var rec *hitrecord.HitRecord
	var obj Hitable
	var hitAnything bool

	closestSoFar := tMax
	epsilon := float32(1e-8)

	for _, h := range hs.hitables {
		if tempRec, _, ok := h.Hit(r, tMin, closestSoFar); ok {
			obj = h
			rec = tempRec
			hitAnything = ok
			closestSoFar = rec.T()
		}
	}

	if hitAnything {
		return obj.HitEdge(r, tMin, closestSoFar+epsilon)
	}

	return nil, false, false
}

func (hs *HitableSlice) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
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

func (hs *HitableSlice) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	weight := 1.0 / float32(len(hs.hitables))
	sum := float32(0)
	for _, h := range hs.hitables {
		sum += weight * h.PDFValue(o, v)
	}
	return sum
}

func (hs *HitableSlice) Random(o *vec3.Vec3Impl, random *fastrandom.LCG) *vec3.Vec3Impl {
	index := int(random.Float32() * float32(len(hs.hitables)))
	return hs.hitables[index].Random(o, random)
}

func (hs *HitableSlice) IsEmitter() bool {
	return false
}
