package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Box)(nil)

// Box represents a box.
type Box struct {
	sides HitableSlice
	pMin  *vec3.Vec3Impl
	pMax  *vec3.Vec3Impl
}

func NewBox(p0 *vec3.Vec3Impl, p1 *vec3.Vec3Impl, mat material.Material) *Box {
	pMin := p0
	pMax := p1

	box := []Hitable{
		NewXYRect(p0.X, p1.X, p0.Y, p1.Y, p1.Z, mat),
		NewFlipNormals(NewXYRect(p0.X, p1.X, p0.Y, p1.Y, p0.Z, mat)),
		NewXZRect(p0.X, p1.X, p0.Z, p1.Z, p1.Y, mat),
		NewFlipNormals(NewXZRect(p0.X, p1.X, p0.Z, p1.Z, p0.Y, mat)),
		NewYZRect(p0.Y, p1.Y, p0.Z, p1.Z, p1.X, mat),
		NewFlipNormals(NewYZRect(p0.Y, p1.Y, p0.Z, p1.Z, p0.X, mat)),
	}

	return &Box{
		sides: *NewSlice(box),
		pMin:  pMin,
		pMax:  pMax,
	}
}

func (b *Box) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	return b.sides.Hit(r, tMin, tMax)
}

func (b *Box) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return b.sides.BoundingBox(time0, time1)
}
