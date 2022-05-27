package hitable

import (
	"math"
	"math/rand"

	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/mat3"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Triangle)(nil)

//Triangle represents a single triangle in the 3d world.
type Triangle struct {
	// Vertices
	vertex0 *vec3.Vec3Impl
	vertex1 *vec3.Vec3Impl
	vertex2 *vec3.Vec3Impl
	// Edges
	edge1 *vec3.Vec3Impl
	edge2 *vec3.Vec3Impl
	// Normal
	normal *vec3.Vec3Impl
	// Used for normal mapping
	tangent   *vec3.Vec3Impl
	bitangent *vec3.Vec3Impl
	// Area
	area float64
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

	return NewTriangleWithUVAndNormal(vertex0, vertex1, vertex2,
		normal, u0, v0, u1, v1, u2, v2, mat)
}

// NewTriangleWithUV returns a new texture triangle.
func NewTriangleWithUVAndNormal(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	normal *vec3.Vec3Impl, u0, v0, u1, v1, u2, v2 float64, mat material.Material) *Triangle {

	deltaU1 := u1 - u0
	deltaU2 := u2 - u0
	deltaV1 := v1 - v0
	deltaV2 := v2 - v0

	edge1 := vec3.Sub(vertex1, vertex0)
	edge2 := vec3.Sub(vertex2, vertex0)

	n := vec3.Cross(edge1, edge2)
	area := n.Length() / 2.0

	f := 1.0 / (deltaU1*deltaV2 - deltaU2*deltaV1)
	tanget := &vec3.Vec3Impl{
		X: f * (deltaV2*edge1.X - deltaV1*edge2.X),
		Y: f * (deltaV2*edge1.Y - deltaV1*edge2.Y),
		Z: f * (deltaV2*edge1.Z - deltaV1*edge2.Z),
	}

	tanget.MakeUnitVector()

	bitangent := &vec3.Vec3Impl{
		X: f * (-deltaU2*edge1.X + deltaU1*edge2.X),
		Y: f * (-deltaU2*edge1.Y + deltaU1*edge2.Y),
		Z: f * (-deltaU2*edge1.Z + deltaU1*edge2.Z),
	}

	bitangent.MakeUnitVector()

	delta := &vec3.Vec3Impl{X: 0.0001, Y: 0.0001, Z: 0.0001}
	min := vec3.Sub(vec3.Min3(vertex0, vertex1, vertex2), delta)
	max := vec3.Add(vec3.Max3(vertex0, vertex1, vertex2), delta)

	return &Triangle{
		vertex0:   vertex0,
		vertex1:   vertex1,
		vertex2:   vertex2,
		edge1:     edge1,
		edge2:     edge2,
		normal:    normal,
		tangent:   tanget,
		bitangent: bitangent,
		area:      area,
		material:  mat,
		u0:        u0,
		u1:        u1,
		u2:        u2,
		v0:        v0,
		v1:        v1,
		v2:        v2,
		bb:        aabb.New(min, max),
	}
}

func (tri *Triangle) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	// https://en.wikipedia.org/wiki/Möller–Trumbore_intersection_algorithm
	epsilon := math.Nextafter(1, 2) - 1

	h := vec3.Cross(r.Direction(), tri.edge2)
	a := vec3.Dot(tri.edge1, h)

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
	q := vec3.Cross(s, tri.edge1)
	v := f * vec3.Dot(r.Direction(), q)
	if v < 0.0 || u+v > 1.0 {
		return nil, nil, false
	}

	t := f * vec3.Dot(tri.edge2, q)
	if t <= epsilon {
		return nil, nil, false
	}

	uv := 1.0 - u - v
	uu := uv*tri.u0 + u*tri.u1 + v*tri.u2
	vv := uv*tri.v0 + u*tri.v1 + v*tri.v2

	normalMap := tri.material.NormalMap()
	if normalMap == nil {
		return hitrecord.New(t, uu, vv, r.PointAtParameter(t), tri.normal), tri.material, true
	}

	// We use OpenGL normal maps.
	normalTangentSpace := normalMap.Value(uu, vv, nil)
	normalTangentSpace.X = 2*normalTangentSpace.X - 1.0
	normalTangentSpace.Y = 2*normalTangentSpace.Y - 1.0
	normalTangentSpace.Z = 2*normalTangentSpace.Z - 1.0

	tbn := mat3.NewTBN(tri.tangent, tri.bitangent, tri.normal)
	normal := mat3.MatrixVectorMul(tbn, normalTangentSpace)
	return hitrecord.New(t, uu, vv, r.PointAtParameter(t), normal), tri.material, true
}

func (tri *Triangle) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return tri.bb, true
}

func (tri *Triangle) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	r := ray.New(o, v, 0)
	if rec, _, ok := tri.Hit(r, 0.001, math.MaxFloat64); ok {
		distanceSquared := rec.T() * rec.T() * v.SquaredLength()
		cosine := math.Abs(vec3.Dot(v, vec3.ScalarDiv(rec.Normal(), v.Length())))
		return distanceSquared / (cosine * tri.area)
	}

	return 0
}

func (tri *Triangle) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	r := rand.Float64()
	randomPoint := &vec3.Vec3Impl{
		X: tri.vertex0.X + r*(tri.vertex1.X-tri.vertex0.X) + (1-r)*(tri.vertex2.X-tri.vertex0.X),
		Y: tri.vertex0.Y + r*(tri.vertex1.Y-tri.vertex0.Y) + (1-r)*(tri.vertex2.Y-tri.vertex0.Y),
		Z: tri.vertex0.Z + r*(tri.vertex1.Z-tri.vertex0.Z) + (1-r)*(tri.vertex2.Z-tri.vertex0.Z),
	}

	return vec3.Sub(randomPoint, o)
}

func (tri *Triangle) IsEmitter() bool {
	return tri.material.IsEmitter()
}

func (tri *Triangle) Vertex0() vec3.Vec3Impl {
	return *tri.vertex0
}

func (tri *Triangle) Vertex1() vec3.Vec3Impl {
	return *tri.vertex1
}

func (tri *Triangle) Vertex2() vec3.Vec3Impl {
	return *tri.vertex2
}

func (tri *Triangle) U0() float64 {
	return tri.u0
}

func (tri *Triangle) U1() float64 {
	return tri.u1
}

func (tri *Triangle) U2() float64 {
	return tri.u2
}

func (tri *Triangle) V0() float64 {
	return tri.v0
}

func (tri *Triangle) V1() float64 {
	return tri.v1
}

func (tri *Triangle) V2() float64 {
	return tri.v2
}

func (tri *Triangle) Material() material.Material {
	return tri.material
}
