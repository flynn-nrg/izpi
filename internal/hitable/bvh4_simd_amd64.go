//go:build amd64

package hitable

// rayAABB4_SIMD_impl is the AMD64 AVX2 implementation
// Implemented in bvh4_avx2.s using 256-bit YMM registers
//
// AVX2 provides true 4-wide SIMD parallelism:
// - YMM registers can hold 8x float32 (we use lower 4)
// - VBROADCASTSS efficiently replicates scalars
// - Parallel min/max/multiply operations
// - VMOVMSKPS extracts comparison mask to integer
//
// Performance expectations:
// - ~10-20% faster than pure Go (bypassing loop overhead)
// - Could bring BVH4 from 3m38s down to ~3min or less
// - No CGO overhead (direct assembly)
//
// Requires: AVX2 support (all modern AMD64 CPUs since ~2013)
//
//go:noescape
func rayAABB4_SIMD_impl(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[4]float32,
	maxX, maxY, maxZ *[4]float32,
	tMax float32,
) uint8

