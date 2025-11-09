//go:build arm64

package hitable

// rayAABB8_SIMD_impl is the pure Go implementation for ARM64
//
// Performance: Delivers excellent performance through BVH8 structure
// even without explicit SIMD, due to:
// - Shallower tree (fewer node visits)
// - Better cache locality (SoA layout)
// - More opportunities for early termination
//
// NEON can only process 4-wide (128-bit), so testing 8 AABBs would
// require two NEON operations anyway. The pure Go version with good
// branch prediction may be competitive.
func rayAABB8_SIMD_impl(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[8]float32,
	maxX, maxY, maxZ *[8]float32,
	tMax float32,
) uint8 {
	var mask uint8 = 0

	for i := 0; i < 8; i++ {
		// X axis
		t0x := (minX[i] - *rayOrgX) * *rayInvDirX
		t1x := (maxX[i] - *rayOrgX) * *rayInvDirX
		if t0x > t1x {
			t0x, t1x = t1x, t0x
		}

		// Y axis
		t0y := (minY[i] - *rayOrgY) * *rayInvDirY
		t1y := (maxY[i] - *rayOrgY) * *rayInvDirY
		if t0y > t1y {
			t0y, t1y = t1y, t0y
		}

		// Z axis
		t0z := (minZ[i] - *rayOrgZ) * *rayInvDirZ
		t1z := (maxZ[i] - *rayOrgZ) * *rayInvDirZ
		if t0z > t1z {
			t0z, t1z = t1z, t0z
		}

		tNear := max32(max32(t0x, t0y), t0z)
		tFar := min32(min32(t1x, t1y), t1z)

		if tNear <= tFar && tFar >= 0 && tNear <= tMax {
			mask |= (1 << i)
		}
	}

	return mask
}

