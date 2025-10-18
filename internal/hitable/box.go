package hitable

import (
	"github.com/flynn-nrg/izpi/internal/aabb"
	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Box)(nil)

// Box represents a box.
type Box struct {
	sides HitableSlice
	mat   material.Material
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
		mat:   mat,
		pMin:  pMin,
		pMax:  pMax,
	}
}

func (b *Box) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	return b.sides.Hit(r, tMin, tMax)
}

func (b *Box) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	return b.sides.HitEdge(r, tMin, tMax)
}

func (b *Box) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return b.sides.BoundingBox(time0, time1)
}

func (b *Box) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	return 0.0
}

func (b *Box) Random(o *vec3.Vec3Impl, _ *fastrandom.LCG) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}

func (b *Box) IsEmitter() bool {
	return b.mat.IsEmitter()
}
