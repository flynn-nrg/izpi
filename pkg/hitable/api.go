// Package hitable implements the methods used to compute intersections between a ray and geometry.
package hitable

import (
	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Hitable defines the methods to compute ray/geometry operations.
type Hitable interface {
	Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool)
	HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool)
	BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool)
	PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64
	Random(o *vec3.Vec3Impl) *vec3.Vec3Impl
	IsEmitter() bool
}
