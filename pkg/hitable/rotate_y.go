package hitable

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*RotateY)(nil)

// RotateY represents a rotation along the Y axis.
type RotateY struct {
	sinTheta float64
	cosTheta float64
	hitable  Hitable
	bbox     *aabb.AABB
	hasBox   bool
}

// NewRotateY returns a new hitable rotated along the Y axis.
func NewRotateY(hitable Hitable, angle float64) *RotateY {
	radians := (math.Pi / 180.0) * angle
	sinTheta := math.Sin(radians)
	cosTheta := math.Cos(radians)
	bbox, hasBox := hitable.BoundingBox(0, 1)
	min := &vec3.Vec3Impl{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64}
	max := &vec3.Vec3Impl{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64}

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float64(i)*bbox.Max().X + (1.0-float64(i))*bbox.Min().X
				y := float64(j)*bbox.Max().Y + (1.0-float64(j))*bbox.Min().Y
				z := float64(k)*bbox.Max().Z + (1.0-float64(k))*bbox.Min().Z
				newx := cosTheta*x + sinTheta*z
				newz := -sinTheta*x + cosTheta*z
				tester := &vec3.Vec3Impl{X: newx, Y: y, Z: newz}

				if tester.X > max.X {
					max.X = tester.X
				}

				if tester.Y > max.Y {
					max.Y = tester.Y
				}

				if tester.Z > max.Z {
					max.Z = tester.Z
				}

				if tester.X < min.X {
					min.X = tester.X
				}

				if tester.Y < min.Y {
					min.Y = tester.Y
				}

				if tester.Z < min.Z {
					min.Z = tester.Z
				}
			}
		}
	}

	return &RotateY{
		sinTheta: sinTheta,
		cosTheta: cosTheta,
		hitable:  hitable,
		bbox:     aabb.New(min, max),
		hasBox:   hasBox,
	}
}

func (ry *RotateY) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	origin := &vec3.Vec3Impl{
		X: ry.cosTheta*r.Origin().X - ry.sinTheta*r.Origin().Z,
		Y: r.Origin().Y,
		Z: ry.sinTheta*r.Origin().X + ry.cosTheta*r.Origin().Z,
	}
	direction := &vec3.Vec3Impl{
		X: ry.cosTheta*r.Direction().X - ry.sinTheta*r.Direction().Z,
		Y: r.Direction().Y,
		Z: ry.sinTheta*r.Direction().X + ry.cosTheta*r.Direction().Z,
	}

	rotatedRay := ray.New(origin, direction, r.Time())

	if hr, mat, ok := ry.hitable.Hit(rotatedRay, tMin, tMax); ok {
		p := &vec3.Vec3Impl{
			X: ry.cosTheta*hr.P().X + ry.sinTheta*hr.P().Z,
			Y: hr.P().Y,
			Z: -ry.sinTheta*hr.P().X + ry.cosTheta*hr.P().Z,
		}
		normal := &vec3.Vec3Impl{
			X: ry.cosTheta*hr.Normal().X + ry.sinTheta*hr.Normal().Z,
			Y: hr.Normal().Y,
			Z: -ry.sinTheta*hr.Normal().X + ry.cosTheta*hr.Normal().Z,
		}

		return hitrecord.New(hr.T(), hr.U(), hr.V(), p, normal), mat, true
	}

	return nil, nil, false
}

func (ry *RotateY) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
	origin := &vec3.Vec3Impl{
		X: ry.cosTheta*r.Origin().X - ry.sinTheta*r.Origin().Z,
		Y: r.Origin().Y,
		Z: ry.sinTheta*r.Origin().X + ry.cosTheta*r.Origin().Z,
	}
	direction := &vec3.Vec3Impl{
		X: ry.cosTheta*r.Direction().X - ry.sinTheta*r.Direction().Z,
		Y: r.Direction().Y,
		Z: ry.sinTheta*r.Direction().X + ry.cosTheta*r.Direction().Z,
	}

	rotatedRay := ray.New(origin, direction, r.Time())

	if hr, ok, edgeOk := ry.hitable.HitEdge(rotatedRay, tMin, tMax); ok {
		p := &vec3.Vec3Impl{
			X: ry.cosTheta*hr.P().X + ry.sinTheta*hr.P().Z,
			Y: hr.P().Y,
			Z: -ry.sinTheta*hr.P().X + ry.cosTheta*hr.P().Z,
		}
		normal := &vec3.Vec3Impl{
			X: ry.cosTheta*hr.Normal().X + ry.sinTheta*hr.Normal().Z,
			Y: hr.Normal().Y,
			Z: -ry.sinTheta*hr.Normal().X + ry.cosTheta*hr.Normal().Z,
		}

		return hitrecord.New(hr.T(), hr.U(), hr.V(), p, normal), true, edgeOk
	}

	return nil, false, false
}

func (ry *RotateY) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return ry.bbox, ry.hasBox
}

func (ry *RotateY) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	return ry.hitable.PDFValue(o, v)
}

func (ry *RotateY) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	return ry.hitable.Random(o)
}

func (ry *RotateY) IsEmitter() bool {
	return ry.hitable.IsEmitter()
}
