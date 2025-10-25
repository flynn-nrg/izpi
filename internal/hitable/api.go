// Package hitable implements the methods used to compute intersections between a ray and geometry.
package hitable

import (
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Hitable defines the methods to compute ray/geometry operations.
type Hitable interface {
	Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool)
	HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool)
	BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool)
	PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float64
	Random(o vec3.Vec3Impl, random *fastrandom.LCG) vec3.Vec3Impl
	IsEmitter() bool
}
