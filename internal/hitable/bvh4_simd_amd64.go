//go:build amd64

package hitable

import "simd/archsimd"

// RayAABB4_SIMD is the AMD64 SSE/AVX implementation using Go 1.26 SIMD intrinsics.
// Implemented using the simd/archsimd package with 128-bit Float32x4 vectors.
//
// Uses 128-bit SSE/AVX operations for optimal 4-wide parallelism:
// - Float32x4 vectors hold exactly 4 float32 (perfect match for 4 AABBs)
// - BroadcastFloat32x4 replicates scalars efficiently
// - Parallel min/max/multiply operations
// - ToBits() extracts 4-bit comparison mask
// - No memory copy overhead (loads directly from [4]float32)
//
// Performance expectations:
// - Should match or exceed assembly version with inlining benefits
// - Eliminates call overhead through inlining
// - Compiler can optimize across function boundaries
//
// Requires: SSE/AVX support (all modern AMD64 CPUs)
//
//	GOEXPERIMENT=simd to enable intrinsics
//
//go:inline
func RayAABB4_SIMD(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[4]float32,
	maxX, maxY, maxZ *[4]float32,
	tMaxParam float32,
) uint8 {
	// Broadcast ray origin components to all 4 lanes
	orgX := archsimd.BroadcastFloat32x4(*rayOrgX)
	orgY := archsimd.BroadcastFloat32x4(*rayOrgY)
	orgZ := archsimd.BroadcastFloat32x4(*rayOrgZ)

	// Broadcast ray inverse direction
	invDirX := archsimd.BroadcastFloat32x4(*rayInvDirX)
	invDirY := archsimd.BroadcastFloat32x4(*rayInvDirY)
	invDirZ := archsimd.BroadcastFloat32x4(*rayInvDirZ)

	// Load AABB bounds directly (no copy needed - perfect fit!)
	minXVec := archsimd.LoadFloat32x4(minX)
	minYVec := archsimd.LoadFloat32x4(minY)
	minZVec := archsimd.LoadFloat32x4(minZ)
	maxXVec := archsimd.LoadFloat32x4(maxX)
	maxYVec := archsimd.LoadFloat32x4(maxY)
	maxZVec := archsimd.LoadFloat32x4(maxZ)

	// ====== X AXIS ======
	// Compute t0x = (minX - orgX) * invDirX
	t0x := minXVec.Sub(orgX).Mul(invDirX)

	// Compute t1x = (maxX - orgX) * invDirX
	t1x := maxXVec.Sub(orgX).Mul(invDirX)

	// Get tNearX = min(t0x, t1x) and tFarX = max(t0x, t1x)
	tNearX := t0x.Min(t1x)
	tFarX := t0x.Max(t1x)

	// Initialize t_min = tNearX, t_max = tFarX
	tMin := tNearX
	tMaxVec := tFarX

	// ====== Y AXIS ======
	t0y := minYVec.Sub(orgY).Mul(invDirY)
	t1y := maxYVec.Sub(orgY).Mul(invDirY)

	tNearY := t0y.Min(t1y)
	tFarY := t0y.Max(t1y)

	// Update t_min = max(t_min, tNearY), t_max = min(t_max, tFarY)
	tMin = tMin.Max(tNearY)
	tMaxVec = tMaxVec.Min(tFarY)

	// ====== Z AXIS ======
	t0z := minZVec.Sub(orgZ).Mul(invDirZ)
	t1z := maxZVec.Sub(orgZ).Mul(invDirZ)

	tNearZ := t0z.Min(t1z)
	tFarZ := t0z.Max(t1z)

	// Final t_min and t_max
	tMin = tMin.Max(tNearZ)
	tMaxVec = tMaxVec.Min(tFarZ)

	// ====== CULLING CHECKS ======
	// Check 1: t_max >= t_min
	cond1 := tMaxVec.GreaterEqual(tMin)

	// Check 2: t_max >= 0.0
	zero := archsimd.BroadcastFloat32x4(0.0)
	cond2 := tMaxVec.GreaterEqual(zero)

	// Check 3: t_min <= tMaxParam (equivalent to tMaxParam >= t_min)
	tMaxScalar := archsimd.BroadcastFloat32x4(tMaxParam)
	cond3 := tMaxScalar.GreaterEqual(tMin)

	// Combine all conditions with AND
	finalMask := cond1.And(cond2).And(cond3)

	// Extract mask to 4-bit integer
	// ToBits() uses MOVMSKPS instruction
	maskBits := finalMask.ToBits()

	// Return 4-bit mask
	return maskBits & 0x0F
}
