package hitable

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*RotateY)(nil)

// RotateY represents a rotation along the Y axis.
type RotateY struct {
	sinTheta float32
	cosTheta float32
	hitable  Hitable
	bbox     *aabb.AABB
	hasBox   bool
}

// NewRotateY returns a new hitable rotated along the Y axis.
func NewRotateY(hitable Hitable, angle float32) *RotateY {
	radians := (math.Pi / 180.0) * angle
	sinTheta := float32(math.Sin(float64(radians)))
	cosTheta := float32(math.Cos(float64(radians)))
	bbox, hasBox := hitable.BoundingBox(0, 1)
	min := &vec3.Vec3Impl{X: math.MaxFloat32, Y: math.MaxFloat32, Z: math.MaxFloat32}
	max := &vec3.Vec3Impl{X: -math.MaxFloat32, Y: -math.MaxFloat32, Z: -math.MaxFloat32}

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				x := float32(i)*bbox.Max().X + (1.0-float32(i))*bbox.Min().X
				y := float32(j)*bbox.Max().Y + (1.0-float32(j))*bbox.Min().Y
				z := float32(k)*bbox.Max().Z + (1.0-float32(k))*bbox.Min().Z
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

func (ry *RotateY) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
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

func (ry *RotateY) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
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

func (ry *RotateY) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return ry.bbox, ry.hasBox
}

func (ry *RotateY) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	return ry.hitable.PDFValue(o, v)
}

func (ry *RotateY) Random(o *vec3.Vec3Impl, random *fastrandom.LCG) *vec3.Vec3Impl {
	return ry.hitable.Random(o, random)
}

func (ry *RotateY) IsEmitter() bool {
	return ry.hitable.IsEmitter()
}
