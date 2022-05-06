package hitable

import (
	"math"

	"gitlab.com/flynn-nrg/izpi/pkg/aabb"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Triangle)(nil)

//Triangle represents a single triangle in the 3d world.
type Triangle struct {
	// Vertices
	vertex0 *vec3.Vec3Impl
	vertex1 *vec3.Vec3Impl
	vertex2 *vec3.Vec3Impl
	// Normal
	normal *vec3.Vec3Impl
	// Material
	material material.Material
	// Texture coordinates
	u0 float64
	u1 float64
	u2 float64
	v0 float64
	v1 float64
	v2 float64
	// Bounding box
	bb *aabb.AABB
}

// NewTriangle returns a new untextured triangle.
func NewTriangle(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	mat material.Material) *Triangle {
	return NewTriangleWithUV(vertex0, vertex1, vertex2, 0, 0, 0, 0, 0, 0, mat)
}

// NewTriangleWithUV returns a new texture triangle.
func NewTriangleWithUV(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	u0, v0, u1, v1, u2, v2 float64, mat material.Material) *Triangle {
	edge1 := vec3.Sub(vertex1, vertex0)
	edge2 := vec3.Sub(vertex2, vertex0)

	normal := vec3.Cross(edge1, edge2)
	normal.MakeUnitVector()

	delta := &vec3.Vec3Impl{X: 0.0001, Y: 0.0001, Z: 00001}
	min := vec3.Sub(vec3.Min3(vertex0, vertex1, vertex2), delta)
	max := vec3.Add(vec3.Max3(vertex0, vertex1, vertex2), delta)

	return &Triangle{
		vertex0:  vertex0,
		vertex1:  vertex1,
		vertex2:  vertex2,
		normal:   normal,
		material: mat,
		u0:       u0,
		u1:       u1,
		u2:       u2,
		v0:       v0,
		v1:       v1,
		v2:       v2,
		bb:       aabb.New(min, max),
	}
}

// NewTriangleWithUV returns a new texture triangle.
func NewTriangleWithUVAndNormal(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	normal *vec3.Vec3Impl, u0, v0, u1, v1, u2, v2 float64, mat material.Material) *Triangle {

	return &Triangle{
		vertex0:  vertex0,
		vertex1:  vertex1,
		vertex2:  vertex2,
		normal:   normal,
		material: mat,
		u0:       u0,
		u1:       u1,
		u2:       u2,
		v0:       v0,
		v1:       v1,
		v2:       v2,
		bb:       aabb.New(vec3.Min3(vertex0, vertex1, vertex2), vec3.Max3(vertex0, vertex1, vertex2)),
	}
}

func (tri *Triangle) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	// https://en.wikipedia.org/wiki/Möller–Trumbore_intersection_algorithm
	epsilon := math.Nextafter(1, 2) - 1

	edge1 := vec3.Sub(tri.vertex1, tri.vertex0)
	edge2 := vec3.Sub(tri.vertex2, tri.vertex0)

	h := vec3.Cross(r.Direction(), edge2)
	a := vec3.Dot(edge1, h)

	if a > -epsilon && a < epsilon {
		// Ray is parallel to triangle.
		return nil, nil, false
	}

	f := 1.0 / a
	s := vec3.Sub(r.Origin(), tri.vertex0)
	u := f * vec3.Dot(s, h)
	if u < 0.0 || u > 1.0 {
		return nil, nil, false
	}
	q := vec3.Cross(s, edge1)
	v := f * vec3.Dot(r.Direction(), q)
	if v < 0.0 || u+v > 1.0 {
		return nil, nil, false
	}

	t := f * vec3.Dot(edge2, q)
	if t <= epsilon {
		return nil, nil, false
	}

	uv := 1.0 - u - v
	uu := uv*tri.u0 + u*tri.u1 + v*tri.u2
	vv := uv*tri.v0 + u*tri.v1 + v*tri.v2

	return hitrecord.New(t, uu, vv, r.PointAtParameter(t), tri.normal), tri.material, true

}

func (tri *Triangle) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return tri.bb, true
}

func (tri *Triangle) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	return 0.0
}

func (tri *Triangle) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}

func (tri *Triangle) IsEmitter() bool {
	return tri.material.IsEmitter()
}
