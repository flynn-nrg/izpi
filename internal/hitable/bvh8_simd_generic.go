//go:build !arm64 && !amd64

package hitable

// rayAABB8_SIMD_impl is the pure Go fallback implementation for BVH8
// Tests a ray against 8 AABBs, returns an 8-bit mask
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

