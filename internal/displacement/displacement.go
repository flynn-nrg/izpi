// Package displacement implements functions to apply displacement maps to triangles and meshes.
package displacement

import (
	"errors"
	"math"
	"time"

	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/mat3"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"

	log "github.com/sirupsen/logrus"
)

// minimalTriangle is a more lightweight structure used when performing multiple tesselation passes.
type minimalTriangle struct {
	// Vertices
	vertex0 vec3.Vec3Impl
	vertex1 vec3.Vec3Impl
	vertex2 vec3.Vec3Impl
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
	a := *vec3.ScalarDiv(vec3.Add(&in.vertex0, &in.vertex1), 2.0)
	b := *vec3.ScalarDiv(vec3.Add(&in.vertex1, &in.vertex2), 2.0)
	c := *vec3.ScalarDiv(vec3.Add(&in.vertex2, &in.vertex0), 2.0)

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

// getDisplacementVariation calculates the variation in displacement values at the triangle vertices
func getDisplacementVariation(tri *minimalTriangle, displacementMap texture.Texture) float64 {
	// Sample displacement at each vertex (we only care about the Z component)
	d0 := displacementMap.Value(tri.u0, tri.v0, nil).Z
	d1 := displacementMap.Value(tri.u1, tri.v1, nil).Z
	d2 := displacementMap.Value(tri.u2, tri.v2, nil).Z

	// Calculate range (max - min) of displacement values
	minDisp := math.Min(d0, math.Min(d1, d2))
	maxDisp := math.Max(d0, math.Max(d1, d2))

	return maxDisp - minDisp
}

func isTessellatedEnough(tri *minimalTriangle, maxDeltaU float64, maxDeltaV float64, displacementMap texture.Texture, min, max, adaptiveThreshold float64) bool {
	// Check UV coordinate constraints (minimum tessellation)
	uvTessellated := math.Abs(tri.u1-tri.u0) <= maxDeltaU &&
		math.Abs(tri.u2-tri.u1) <= maxDeltaU &&
		math.Abs(tri.u0-tri.u2) <= maxDeltaU &&
		math.Abs(tri.v1-tri.v0) <= maxDeltaV &&
		math.Abs(tri.v2-tri.v1) <= maxDeltaV &&
		math.Abs(tri.v0-tri.v2) <= maxDeltaV

	// If UV tessellation is not satisfied, we must subdivide
	if !uvTessellated {
		return false
	}

	// Check displacement variation (adaptive tessellation)
	// If the displacement variation is small, we can stop tessellating
	variation := getDisplacementVariation(tri, displacementMap)
	displacementRange := math.Abs(max - min)

	// Normalize variation by the displacement range
	normalizedVariation := variation * displacementRange

	// Stop tessellating if variation is below threshold
	return normalizedVariation <= adaptiveThreshold
}

// ApplyDisplacementMap tessellates the triangles and applies the displacement map to all of them.
// It uses adaptive tessellation that only subdivides areas with significant displacement variation.
func ApplyDisplacementMap(triangles []*hitable.Triangle, displacementMap texture.Texture, min, max float64) ([]*hitable.Triangle, error) {
	var resU, resV int

	if displacementTexture, ok := displacementMap.(*texture.ImageTxt); !ok {
		return nil, errors.New("only ImageTxt texture type is supported")
	} else {
		resU = displacementTexture.SizeX()
		resV = displacementTexture.SizeY()
	}

	in := []*minimalTriangle{}

	for _, tri := range triangles {
		in = append(in, &minimalTriangle{
			vertex0:  tri.Vertex0(),
			vertex1:  tri.Vertex1(),
			vertex2:  tri.Vertex2(),
			material: tri.Material(),
			u0:       tri.U0(),
			u1:       tri.U1(),
			u2:       tri.U2(),
			v0:       tri.V0(),
			v1:       tri.V1(),
			v2:       tri.V2(),
		})
	}

	// Use coarser UV tessellation limits (4x the texture pixel size)
	// This allows adaptive tessellation to take over for detail
	tessellationFactor := 4.0
	maxDeltaU := tessellationFactor / float64(resU-1)
	maxDeltaV := tessellationFactor / float64(resV-1)

	// Adaptive threshold: stop tessellating when displacement variation is less than this
	// This is in world space units. Smaller values = more tessellation for finer detail.
	// For water surfaces, we can use a higher threshold since smooth waves don't need extreme detail
	adaptiveThreshold := 2.0 // Adjust this based on your scene scale and desired quality

	tessellated := applyTessellation(in, maxDeltaU, maxDeltaV, displacementMap, min, max, adaptiveThreshold)

	return applyDisplacement(tessellated, displacementMap, min, max), nil
}

func applyTessellation(in []*minimalTriangle, maxDeltaU float64, maxDeltaV float64, displacementMap texture.Texture, min, max, adaptiveThreshold float64) []*minimalTriangle {
	out := []*minimalTriangle{}

	log.Info("Applying adaptive tessellation")
	log.Infof("Tessellation parameters: UV delta=%.6f, adaptive threshold=%.3f", maxDeltaU, adaptiveThreshold)
	startTime := time.Now()

	iterations := 0
	for {
		if len(in) == 0 {
			log.Infof("Adaptive tessellation completed. Created %v triangles in %v iterations over %v", len(out), iterations, time.Since(startTime))
			return out
		}

		iterations++
		toIn := []*minimalTriangle{}
		for _, triangle := range in {
			newTriangles := tessellate(triangle)
			for _, tessellated := range newTriangles {
				if isTessellatedEnough(tessellated, maxDeltaU, maxDeltaV, displacementMap, min, max, adaptiveThreshold) {
					out = append(out, tessellated)
				} else {
					toIn = append(toIn, tessellated)
				}
			}
		}

		in = toIn
	}
}

func applyDisplacement(in []*minimalTriangle, displacementMap texture.Texture, min, max float64) []*hitable.Triangle {
	displaced := []*hitable.Triangle{}

	log.Info("Applying displacement")
	startTime := time.Now()

	for _, tri := range in {

		edge1 := vec3.Sub(&tri.vertex1, &tri.vertex0)
		edge2 := vec3.Sub(&tri.vertex2, &tri.vertex0)
		normal := vec3.Cross(edge1, edge2)
		normal.MakeUnitVector()
		deltaU1 := tri.u1 - tri.u0
		deltaU2 := tri.u2 - tri.u0
		deltaV1 := tri.v1 - tri.v0
		deltaV2 := tri.v2 - tri.v0

		f := 1.0 / (deltaU1*deltaV2 - deltaU2*deltaV1)
		tangent := &vec3.Vec3Impl{
			X: f * (deltaV2*edge1.X - deltaV1*edge2.X),
			Y: f * (deltaV2*edge1.Y - deltaV1*edge2.Y),
			Z: f * (deltaV2*edge1.Z - deltaV1*edge2.Z),
		}

		tangent.MakeUnitVector()

		bitangent := &vec3.Vec3Impl{
			X: f * (-deltaU2*edge1.X + deltaU1*edge2.X),
			Y: f * (-deltaU2*edge1.Y + deltaU1*edge2.Y),
			Z: f * (-deltaU2*edge1.Z + deltaU1*edge2.Z),
		}

		bitangent.MakeUnitVector()

		tbn := mat3.NewTBN(tangent, bitangent, normal)

		displacementVertex0TangentSpace := displacementMap.Value(tri.u0, tri.v0, nil)
		displacementVertex0TangentSpace.X = 0.0
		displacementVertex0TangentSpace.Y = 0.0
		displacementVertex0TangentSpace.Z = min + ((max - min) * displacementVertex0TangentSpace.Z)
		displacementVertex0 := mat3.MatrixVectorMul(tbn, displacementVertex0TangentSpace)
		vertex0 := vec3.Add(&tri.vertex0, displacementVertex0)

		displacementVertex1TangentSpace := displacementMap.Value(tri.u1, tri.v1, nil)
		displacementVertex1TangentSpace.X = 0.0
		displacementVertex1TangentSpace.Y = 0.0
		displacementVertex1TangentSpace.Z = min + ((max - min) * displacementVertex1TangentSpace.Z)
		displacementVertex1 := mat3.MatrixVectorMul(tbn, displacementVertex1TangentSpace)
		vertex1 := vec3.Add(&tri.vertex1, displacementVertex1)

		displacementVertex2TangentSpace := displacementMap.Value(tri.u2, tri.v2, nil)
		displacementVertex2TangentSpace.X = 0.0
		displacementVertex2TangentSpace.Y = 0.0
		displacementVertex2TangentSpace.Z = min + ((max - min) * displacementVertex2TangentSpace.Z)
		displacementVertex2 := mat3.MatrixVectorMul(tbn, displacementVertex2TangentSpace)
		vertex2 := vec3.Add(&tri.vertex2, displacementVertex2)

		displaced = append(displaced, hitable.NewTriangleWithUV(vertex0, vertex1, vertex2,
			tri.u0, tri.v0, tri.u1, tri.v1, tri.u2, tri.v2, tri.material))
	}

	log.Infof("Displacement completed in %v", time.Since(startTime))

	return displaced
}
