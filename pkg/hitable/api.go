// Package hitable implements the methods used to compute intersections between a ray and geometry.
package hitable

import (
	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
)

// Hitable defines the methods compute ray/geometry operations.
type Hitable interface {
	Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool)
	BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool)
}
