//go:build amd64

package hitable

import "simd/archsimd"

// RayAABB4_SIMD is the AMD64 AVX2 implementation using Go 1.26 SIMD intrinsics.
// Implemented using the simd/archsimd package with 256-bit Float32x8 vectors.
//
// AVX2 provides true 4-wide SIMD parallelism:
// - Float32x8 vectors can hold 8x float32 (we use lower 4 for 4 AABBs)
// - BroadcastFloat32x8 efficiently replicates scalars
// - Parallel min/max/multiply operations
// - ToBits() extracts comparison mask to integer
//
// Performance expectations:
// - ~10-30% faster than pure Go (inlining + no call overhead)
// - Could bring BVH4 from 3m38s down to ~3min or less
// - Better than assembly: enables inlining and compiler optimizations
//
// Requires: AVX2 support (all modern AMD64 CPUs since ~2013)
//          GOEXPERIMENT=simd to enable intrinsics
//
//go:inline
func RayAABB4_SIMD(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[4]float32,
	maxX, maxY, maxZ *[4]float32,
	tMaxParam float32,
) uint8 {
	// Broadcast ray origin components to all 8 lanes (we'll use lower 4)
	orgX := archsimd.BroadcastFloat32x8(*rayOrgX)
	orgY := archsimd.BroadcastFloat32x8(*rayOrgY)
	orgZ := archsimd.BroadcastFloat32x8(*rayOrgZ)

	// Broadcast ray inverse direction
	invDirX := archsimd.BroadcastFloat32x8(*rayInvDirX)
	invDirY := archsimd.BroadcastFloat32x8(*rayInvDirY)
	invDirZ := archsimd.BroadcastFloat32x8(*rayInvDirZ)

	// Load AABB bounds (4 float32 values each)
	// We need to create [8]float32 arrays with the 4 values in the lower half
	var minXArray [8]float32
	var minYArray [8]float32
	var minZArray [8]float32
	var maxXArray [8]float32
	var maxYArray [8]float32
	var maxZArray [8]float32

	copy(minXArray[:4], minX[:])
	copy(minYArray[:4], minY[:])
	copy(minZArray[:4], minZ[:])
	copy(maxXArray[:4], maxX[:])
	copy(maxYArray[:4], maxY[:])
	copy(maxZArray[:4], maxZ[:])

	minXVec := archsimd.LoadFloat32x8(&minXArray)
	minYVec := archsimd.LoadFloat32x8(&minYArray)
	minZVec := archsimd.LoadFloat32x8(&minZArray)
	maxXVec := archsimd.LoadFloat32x8(&maxXArray)
	maxYVec := archsimd.LoadFloat32x8(&maxYArray)
	maxZVec := archsimd.LoadFloat32x8(&maxZArray)

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
	zero := archsimd.BroadcastFloat32x8(0.0)
	cond2 := tMaxVec.GreaterEqual(zero)

	// Check 3: t_min <= tMaxParam (equivalent to tMaxParam >= t_min)
	tMaxScalar := archsimd.BroadcastFloat32x8(tMaxParam)
	cond3 := tMaxScalar.GreaterEqual(tMin)

	// Combine all conditions with AND
	finalMask := cond1.And(cond2).And(cond3)

	// Extract mask to integer (lower 4 bits are what we care about)
	// ToBits() uses VMOVMSKPS instruction
	maskBits := finalMask.ToBits()

	// Return lower 4 bits as uint8 (upper 4 bits are for lanes 4-7, which we don't use)
	return maskBits & 0x0F
}
