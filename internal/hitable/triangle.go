package hitable

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/aabb"
	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/mat3"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/segment"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Hitable = (*Triangle)(nil)

// Triangle represents a single triangle in the 3d world.
type Triangle struct {
	// Vertices
	vertex0 *vec3.Vec3Impl
	vertex1 *vec3.Vec3Impl
	vertex2 *vec3.Vec3Impl
	// Vertex normals
	perVertexNormals bool
	vn0              *vec3.Vec3Impl
	vn1              *vec3.Vec3Impl
	vn2              *vec3.Vec3Impl

	// Edges
	edge1 *vec3.Vec3Impl
	edge2 *vec3.Vec3Impl
	// Normal
	normal *vec3.Vec3Impl
	// Used for normal mapping
	tangent   *vec3.Vec3Impl
	bitangent *vec3.Vec3Impl
	// Area
	area float32
	// Material
	material material.Material
	// Texture coordinates
	u0 float32
	u1 float32
	u2 float32
	v0 float32
	v1 float32
	v2 float32
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
	u0, v0, u1, v1, u2, v2 float32, mat material.Material) *Triangle {
	edge1 := vec3.Sub(vertex1, vertex0)
	edge2 := vec3.Sub(vertex2, vertex0)

	normal := vec3.Cross(edge1, edge2)
	normal.MakeUnitVector()

	return NewTriangleWithUVAndNormal(vertex0, vertex1, vertex2,
		normal, u0, v0, u1, v1, u2, v2, mat)
}

// NewTriangleWithUV returns a new texture triangle.
func NewTriangleWithUVAndNormal(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	normal *vec3.Vec3Impl, u0, v0, u1, v1, u2, v2 float32, mat material.Material) *Triangle {

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

	// Calculate bounding box with relative epsilon based on triangle size
	min := vec3.Min3(vertex0, vertex1, vertex2)
	max := vec3.Max3(vertex0, vertex1, vertex2)

	// Calculate size of triangle
	size := vec3.Sub(max, min)
	maxDim := math.Max(size.X, math.Max(size.Y, size.Z))

	// Use epsilon relative to triangle size, with a minimum value
	epsilon := math.Max(maxDim*1e-4, 1e-6)
	delta := &vec3.Vec3Impl{X: epsilon, Y: epsilon, Z: epsilon}

	min = vec3.Sub(min, delta)
	max = vec3.Add(max, delta)

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

// NewTriangleWithUVAndVertexNormals returns a new textured triangle with per vertex normals.
func NewTriangleWithUVAndVertexNormals(vertex0 *vec3.Vec3Impl, vertex1 *vec3.Vec3Impl, vertex2 *vec3.Vec3Impl,
	vn0 *vec3.Vec3Impl, vn1 *vec3.Vec3Impl, vn2 *vec3.Vec3Impl, u0, v0, u1, v1, u2, v2 float32,
	mat material.Material) *Triangle {

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
		vertex0:          vertex0,
		vertex1:          vertex1,
		vertex2:          vertex2,
		perVertexNormals: true,
		vn0:              vec3.UnitVector(vn0),
		vn1:              vec3.UnitVector(vn1),
		vn2:              vec3.UnitVector(vn2),
		edge1:            edge1,
		edge2:            edge2,
		tangent:          tanget,
		bitangent:        bitangent,
		area:             area,
		material:         mat,
		u0:               u0,
		u1:               u1,
		u2:               u2,
		v0:               v0,
		v1:               v1,
		v2:               v2,
		bb:               aabb.New(min, max),
	}
}

func (tri *Triangle) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	// https://en.wikipedia.org/wiki/Möller–Trumbore_intersection_algorithm
	// Use a larger epsilon for better numerical stability
	epsilon := 1e-8

	h := vec3.Cross(r.Direction(), tri.edge2)
	a := vec3.Dot(tri.edge1, h)

	if math.Abs(a) < epsilon {
		// Ray is parallel to triangle.
		return nil, nil, false
	}

	f := 1.0 / a
	s := vec3.Sub(r.Origin(), tri.vertex0)
	u := f * vec3.Dot(s, h)
	if u < -epsilon || u > 1.0+epsilon {
		return nil, nil, false
	}
	q := vec3.Cross(s, tri.edge1)
	v := f * vec3.Dot(r.Direction(), q)
	if v < -epsilon || u+v > 1.0+epsilon {
		return nil, nil, false
	}

	t := f * vec3.Dot(tri.edge2, q)
	if t < tMin || t > tMax {
		return nil, nil, false
	}

	// Compute barycentric coordinates with better precision
	w := 1.0 - u - v
	// Normalize barycentric coordinates
	sum := u + v + w
	if math.Abs(sum-1.0) > epsilon {
		u /= sum
		v /= sum
		w /= sum
	}

	// Interpolate UV coordinates
	uu := w*tri.u0 + u*tri.u1 + v*tri.u2
	vv := w*tri.v0 + u*tri.v1 + v*tri.v2

	var normal *vec3.Vec3Impl

	if tri.perVertexNormals {
		// Interpolate vertex normals
		normal0 := vec3.ScalarMul(tri.vn0, w)
		normal1 := vec3.ScalarMul(tri.vn1, u)
		normal2 := vec3.ScalarMul(tri.vn2, v)
		normal = vec3.UnitVector(vec3.Add(normal0, normal1, normal2))
	} else {
		normal = tri.normal
	}

	// Handle normal mapping
	normalMap := tri.material.NormalMap()
	if normalMap == nil {
		return hitrecord.New(t, uu, vv, r.PointAtParameter(t), normal), tri.material, true
	}

	// We use OpenGL normal maps.
	normalTangentSpace := normalMap.Value(uu, vv, nil)
	normalTangentSpace.X = 2*normalTangentSpace.X - 1.0
	normalTangentSpace.Y = 2*normalTangentSpace.Y - 1.0
	normalTangentSpace.Z = 2*normalTangentSpace.Z - 1.0

	tbn := mat3.NewTBN(tri.tangent, tri.bitangent, normal)
	newNormal := mat3.MatrixVectorMul(tbn, normalTangentSpace)
	newNormal.MakeUnitVector() // Ensure the final normal is normalized
	return hitrecord.New(t, uu, vv, r.PointAtParameter(t), newNormal), tri.material, true
}

func (tri *Triangle) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return tri.bb, true
}

func (tri *Triangle) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	r := ray.New(o, v, 0)
	if rec, _, ok := tri.Hit(r, 0.001, math.Maxfloat32); ok {
		distanceSquared := rec.T() * rec.T() * v.SquaredLength()
		cosine := math.Abs(vec3.Dot(v, vec3.ScalarDiv(rec.Normal(), v.Length())))
		return distanceSquared / (cosine * tri.area)
	}

	return 0
}

func (tri *Triangle) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	rec, _, ok := tri.Hit(r, tMin, tMax)
	if !ok {
		return nil, false, false
	}

	segments := []*segment.Segment{
		{A: tri.vertex0, B: tri.vertex1},
		{A: tri.vertex1, B: tri.vertex2},
		{A: tri.vertex2, B: tri.vertex0},
	}

	c := rec.P()
	for _, s := range segments {
		if segment.Belongs(s, c) {
			return rec, true, true
		}
	}

	return rec, true, false
}

/*
func (tri *Triangle) Random(o *vec3.Vec3Impl) *vec3.Vec3Impl {
	r := rand.float32()
	randomPoint := &vec3.Vec3Impl{
		X: tri.vertex0.X + r*(tri.vertex1.X-tri.vertex0.X) + (1-r)*(tri.vertex2.X-tri.vertex0.X),
		Y: tri.vertex0.Y + r*(tri.vertex1.Y-tri.vertex0.Y) + (1-r)*(tri.vertex2.Y-tri.vertex0.Y),
		Z: tri.vertex0.Z + r*(tri.vertex1.Z-tri.vertex0.Z) + (1-r)*(tri.vertex2.Z-tri.vertex0.Z),
	}

	return vec3.Sub(randomPoint, o)
}
*/

func (tri *Triangle) Random(o *vec3.Vec3Impl, random *fastrandom.LCG) *vec3.Vec3Impl {
	t1 := random.float32()
	randomPoint01 := vec3.Lerp(tri.vertex0, tri.vertex1, t1)
	t2 := random.float32()
	randomPoint02 := vec3.Lerp(tri.vertex0, tri.vertex2, t2)
	t3 := random.float32()
	randomPoint := vec3.Lerp(randomPoint01, randomPoint02, t3)

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

func (tri *Triangle) U0() float32 {
	return tri.u0
}

func (tri *Triangle) U1() float32 {
	return tri.u1
}

func (tri *Triangle) U2() float32 {
	return tri.u2
}

func (tri *Triangle) V0() float32 {
	return tri.v0
}

func (tri *Triangle) V1() float32 {
	return tri.v1
}

func (tri *Triangle) V2() float32 {
	return tri.v2
}

func (tri *Triangle) Material() material.Material {
	return tri.material
}
