# BVH4 SIMD Implementation

This directory contains the BVH4 (4-wide BVH) implementation with SIMD-optimized ray-AABB intersection tests.

## Overview

The `RayAABB4_SIMD` function tests a ray against 4 axis-aligned bounding boxes (AABBs) simultaneously,
returning a bitmask indicating which boxes were hit. This is the performance-critical function for
BVH4 traversal in the path tracer.

## Implementations

### AMD64 (x86-64) - Go 1.26 SIMD Intrinsics

**File:** `bvh4_simd_amd64.go`

Uses Go 1.26's new `simd/archsimd` package with AVX2 intrinsics:
- **Float32x8** vectors (256-bit) hold 8 float32 values (we use lower 4 for 4 AABBs)
- **BroadcastFloat32x8** replicates scalar ray parameters to all lanes
- **Sub/Mul/Min/Max** perform parallel arithmetic on all AABBs
- **GreaterEqual** creates comparison masks
- **ToBits()** extracts the result mask using VMOVMSKPS instruction

**Advantages over previous assembly version:**
- Compiler can inline the function (assembly had call overhead)
- Compiler can optimize across function boundaries
- More maintainable and readable than `.s` files
- Performance should be ~10-30% better than pure Go

**Requirements:**
- CPU with AVX2 support (all modern x86-64 CPUs since ~2013)
- Build with `GOEXPERIMENT=simd` environment variable

### ARM64 (Apple Silicon, ARM servers) - Pure Go

**File:** `bvh4_simd_arm64.go`

Uses a pure Go loop-based implementation:
- Already delivers 2.91x speedup (22m59s → 7m54s) on M2 Max
- Compiler does an excellent job optimizing this loop structure
- CGO + NEON intrinsics would actually be **2.4x slower** (35ns vs 15ns) due to call overhead

**Note:** Go 1.26's `simd/archsimd` does NOT yet support ARM64/NEON intrinsics.
ARM64 support is planned for Go 1.27 on the `dev.simd` branch.

When Go 1.27 adds NEON intrinsics, we can create a **Float32x4** version that should:
- Eliminate loop overhead completely
- Reduce branch mispredictions
- Enable further compiler optimizations
- Expected improvement: ~20-40% faster than current pure Go version

### Generic Fallback

**File:** `bvh4_simd_generic.go`

Pure Go implementation for other architectures (RISC-V, MIPS, etc.).
Same loop-based approach as ARM64 version.

## Building and Testing

### Prerequisites

```bash
# Check Go version (need 1.26+)
go version  # Should show go1.26.0 or later
```

### Building on AMD64 (x86-64)

```bash
# Build with SIMD intrinsics enabled
GOEXPERIMENT=simd go build ./internal/hitable

# Run correctness tests
GOEXPERIMENT=simd go test ./internal/hitable -run TestRayAABB4_SIMD_Correctness -v

# Run all SIMD tests
GOEXPERIMENT=simd go test ./internal/hitable -run SIMD -v
```

### Building on ARM64 (Apple Silicon)

```bash
# Uses pure Go implementation (no GOEXPERIMENT needed)
go test ./internal/hitable -run TestRayAABB4_SIMD_Correctness -v

# To cross-compile and test the AMD64 version from ARM64
GOEXPERIMENT=simd GOARCH=amd64 go test -c ./internal/hitable
```

### Benchmarking

```bash
# Benchmark SIMD implementation (on native architecture)
GOEXPERIMENT=simd go test ./internal/hitable -bench BenchmarkRayAABB4 -benchmem -v

# Compare SIMD vs Reference implementation
GOEXPERIMENT=simd go test ./internal/hitable -bench "BenchmarkRayAABB4_(SIMD|Reference)" -benchmem

# Full BVH comparison (BVH2 vs BVH4)
GOEXPERIMENT=simd go test ./internal/hitable -bench BenchmarkBVHComparison -benchtime=10s -v
```

### Expected Performance

#### AMD64 (previous assembly vs new intrinsics)
- **Previous:** Assembly `.s` file with call overhead
- **New:** Inlined intrinsics, expected ~10-30% improvement
- **Reason:** No call overhead + compiler can optimize across boundaries

#### ARM64 (current pure Go)
- **Current:** 2.91x speedup over BVH2 (22m59s → 7m54s) on M2 Max
- **Future (Go 1.27):** With NEON intrinsics, expect additional ~20-40% improvement

## Implementation Notes

### Why Float32 instead of Float64?

The BVH4 implementation uses conservative float32 conversions for AABB bounds:
- `conservativeFloat32Min()` - rounds down to ensure we don't miss intersections
- `conservativeFloat32Max()` - rounds up to ensure we don't miss intersections
- Ray parameters stay in float32 for SIMD efficiency
- This reduces memory footprint (4 bytes vs 8 bytes per value)
- Tests verify no precision-related misses occur

### Data Layout

The function uses Structure-of-Arrays (SoA) layout:
- 6 separate arrays: minX, minY, minZ, maxX, maxY, maxZ
- Each array contains 4 float32 values (one per AABB)
- This enables efficient SIMD loading and processing
- Better cache utilization than Array-of-Structures

### Algorithm

The ray-AABB intersection test uses the slab method:
1. For each axis (X, Y, Z):
   - Compute intersection distances: t0 = (min - org) * invDir, t1 = (max - org) * invDir
   - Swap if needed: tNear = min(t0, t1), tFar = max(t0, t1)
2. Find overlap: tMin = max(tNearX, tNearY, tNearZ), tMax = min(tFarX, tFarY, tFarZ)
3. Check intersection: tMin <= tMax && tMax >= 0 && tMin <= rayTMax

The SIMD version performs all 4 checks in parallel, then extracts a 4-bit mask.

## Migration from Assembly

The previous AMD64 implementation used hand-written AVX2 assembly (`bvh4_avx2.s`).
This has been replaced with Go intrinsics for the following benefits:

| Aspect | Assembly (.s) | Intrinsics (Go 1.26) |
|--------|---------------|----------------------|
| Inlining | ❌ No | ✅ Yes |
| Call overhead | ❌ ~2-5ns | ✅ None (inlined) |
| Optimization | ❌ Manual only | ✅ Compiler optimizes |
| Maintainability | ❌ Hard to read/modify | ✅ Clear Go code |
| Type safety | ❌ No type checking | ✅ Type-safe |
| Performance | ✅ Good | ✅ Better (with inlining) |

## Future Work

### Go 1.27 - ARM64 NEON Intrinsics

When ARM64 support lands in `simd/archsimd` (Go 1.27), we can create:

```go
//go:build arm64

func RayAABB4_SIMD(...) uint8 {
    // Use Float32x4 (NEON 128-bit Q registers)
    orgX := archsimd.BroadcastFloat32x4(*rayOrgX)
    // ... similar to AMD64 but with 4-wide vectors
    
    // NEON provides:
    // - FMUL (multiply), FMIN (min), FMAX (max)
    // - FCMGE (greater-equal comparison)
    // - Efficient lane-wise operations
}
```

Expected performance: ~20-40% faster than current pure Go ARM64 version.

## References

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [simd/archsimd Package Documentation](https://pkg.go.dev/simd/archsimd)
- [SIMD in Go - GitHub Issue #73787](https://go.dev/issue/73787)

