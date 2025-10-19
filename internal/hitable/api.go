// Package hitable implements the methods used to compute intersections between a ray and geometry.
package hitable

import (
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
)

// Hitable defines the methods to compute ray/geometry operations.
type Hitable interface {
	Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool)
	HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool)
	BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool)
	PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32
	Random(o *vec3.Vec3Impl, random *fastrandom.XorShift) *vec3.Vec3Impl
	IsEmitter() bool
}
