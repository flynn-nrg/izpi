//go:build arm64

package hitable

// rayAABB4_SIMD_impl is the pure Go implementation for ARM64
// This provides excellent performance through:
// - Better cache locality from BVH4 structure-of-arrays layout
// - Reduced branch mispredictions from processing 4 AABBs together
// - Potential for future Go compiler auto-vectorization
//
// Note: We tested CGO + ARM NEON intrinsics but found that CGO call overhead
// (35.47 ns/op vs 15.07 ns/op for pure Go) negated any SIMD benefits.
// For tight loops with millions of calls, pure Go is significantly faster.
func rayAABB4_SIMD_impl(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[4]float32,
	maxX, maxY, maxZ *[4]float32,
	tMax float32,
) uint8 {
	var mask uint8 = 0

	for i := 0; i < 4; i++ {
		// Compute intersection distances for X axis
		t0x := (minX[i] - *rayOrgX) * *rayInvDirX
		t1x := (maxX[i] - *rayOrgX) * *rayInvDirX
		if t0x > t1x {
			t0x, t1x = t1x, t0x
		}

		// Compute intersection distances for Y axis
		t0y := (minY[i] - *rayOrgY) * *rayInvDirY
		t1y := (maxY[i] - *rayOrgY) * *rayInvDirY
		if t0y > t1y {
			t0y, t1y = t1y, t0y
		}

		// Compute intersection distances for Z axis
		t0z := (minZ[i] - *rayOrgZ) * *rayInvDirZ
		t1z := (maxZ[i] - *rayOrgZ) * *rayInvDirZ
		if t0z > t1z {
			t0z, t1z = t1z, t0z
		}

		// Find the overlap
		tNear := max32(max32(t0x, t0y), t0z)
		tFar := min32(min32(t1x, t1y), t1z)

		// Check if there's an intersection
		if tNear <= tFar && tFar >= 0 && tNear <= tMax {
			mask |= (1 << i)
		}
	}

	return mask
}

