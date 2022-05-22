// Package displacement implements functions to apply displacement maps to triangles and meshes.
package displacement

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// minimalTriangle is a more lightweight structure used when performing multiple tesselation passes.
type minimalTriangle struct {
	// Vertices
	vertex0 *vec3.Vec3Impl
	vertex1 *vec3.Vec3Impl
	vertex2 *vec3.Vec3Impl
	// Material
	material material.Material
	// Texture coordinates
	u0 float64
	u1 float64
	u2 float64
	v0 float64
	v1 float64
	v2 float64
}

// tessellate splits a triangle in four smaller triangles.
func tessellate(in *minimalTriangle) []*minimalTriangle {
	a := vec3.ScalarDiv(vec3.Add(in.vertex0, in.vertex1), 2.0)
	b := vec3.ScalarDiv(vec3.Add(in.vertex1, in.vertex2), 2.0)
	c := vec3.ScalarDiv(vec3.Add(in.vertex2, in.vertex0), 2.0)

	ua := (in.u0 + in.u1) / 2.0
	va := (in.v0 + in.v1) / 2.0

	ub := (in.u1 + in.u2) / 2.0
	vb := (in.v1 + in.v2) / 2.0

	uc := (in.u2 + in.u0) / 2.0
	vc := (in.v2 + in.v0) / 2.0

	return []*minimalTriangle{
		{
			vertex0:  in.vertex0,
			vertex1:  a,
			vertex2:  c,
			material: in.material,
			u0:       in.u0,
			u1:       ua,
			u2:       uc,
			v0:       in.v0,
			v1:       va,
			v2:       vc,
		},
		{
			vertex0:  a,
			vertex1:  b,
			vertex2:  c,
			material: in.material,
			u0:       ua,
			u1:       ub,
			u2:       uc,
			v0:       va,
			v1:       vb,
			v2:       vc,
		},
		{
			vertex0:  a,
			vertex1:  in.vertex1,
			vertex2:  b,
			material: in.material,
			u0:       ua,
			u1:       in.u1,
			u2:       ub,
			v0:       va,
			v1:       in.v1,
			v2:       vb,
		},
		{
			vertex0:  c,
			vertex1:  b,
			vertex2:  in.vertex2,
			material: in.material,
			u0:       uc,
			u1:       ub,
			u2:       in.u2,
			v0:       vc,
			v1:       vb,
			v2:       in.v2,
		},
	}

}

func requiresFurtherTessellation(tri *minimalTriangle, resU int64, resV int64) bool {
	maxDeltaU := 1.0 / float64(resU)
	maxDeltaV := 1.0 / float64(resV)

	return math.Abs(tri.u1-tri.u0) <= maxDeltaU &&
		math.Abs(tri.u2-tri.u1) <= maxDeltaU &&
		math.Abs(tri.u0-tri.u2) <= maxDeltaU &&
		math.Abs(tri.v1-tri.v0) <= maxDeltaV &&
		math.Abs(tri.v2-tri.v1) <= maxDeltaV &&
		math.Abs(tri.v0-tri.v2) <= maxDeltaV
}

// ApplyDisplacementMap tessellates the triangles and applies the displacement map to all of them.
func ApplyDisplacementMap(triangles []*hitable.Triangle, normalMap texture.Texture) []*hitable.Triangle {
	return nil
}
