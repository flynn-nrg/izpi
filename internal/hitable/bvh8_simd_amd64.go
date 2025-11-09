//go:build amd64

package hitable

// rayAABB8_SIMD_impl is the pure Go implementation for AMD64
//
// Performance note: This tests 8 AABBs sequentially in Go.
// Future optimization: AVX-512 could test all 8 AABBs in parallel
// with zmm registers (512-bit = 16x float32, but we only need 8).
//
// AVX-512 would provide true 8-wide SIMD, but:
// - Requires manual instruction encoding (Go assembler lacks AVX-512)
// - Only beneficial if CGO overhead can be avoided
// - Need to measure if algorithmic benefit justifies complexity
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

