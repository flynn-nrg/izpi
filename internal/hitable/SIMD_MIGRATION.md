# Migration from Assembly to Go 1.26 SIMD Intrinsics

## What Changed

### Before (Go 1.25 and earlier)

AMD64 implementation used hand-written AVX2 assembly in `bvh4_avx2.s`:
- 142 lines of x86-64 assembly
- Manual register management (YMM0-YMM15)
- Call overhead (function not inlineable)
- Required deep knowledge of x86-64 assembly and calling conventions
- Used `//go:noescape` directive

```
// Previous files:
internal/hitable/bvh4_avx2.s          ❌ DELETED
internal/hitable/bvh4_simd_amd64.go   ✅ UPDATED (now uses intrinsics)
```

### After (Go 1.26 with GOEXPERIMENT=simd)

AMD64 implementation uses Go intrinsics from `simd/archsimd` package:
- 120 lines of clean, type-safe Go code
- Compiler manages registers automatically
- **Inlineable** - eliminates call overhead
- Compiler can optimize across function boundaries
- Readable and maintainable

```go
// New implementation in bvh4_simd_amd64.go
func RayAABB4_SIMD(...) uint8 {
    orgX := archsimd.BroadcastFloat32x8(*rayOrgX)
    // ... clean Go code using SIMD intrinsics
}
```

## Performance Expectations

### AMD64 (Intel/AMD x86-64)

| Metric | Assembly (old) | Intrinsics (new) | Expected Change |
|--------|----------------|------------------|-----------------|
| Inlining | No | Yes | +10-20% |
| Call overhead | ~2-5ns | 0ns (inlined) | Eliminated |
| Compiler optimizations | Limited | Full | +5-10% |
| **Total expected gain** | Baseline | **+15-30% faster** | 🚀 |

**Previous benchmark baseline:**
- BVH4 render time: 3m38s on AMD64

**Expected with intrinsics:**
- BVH4 render time: ~2m45s - 3m05s on AMD64

### ARM64 (Apple Silicon)

**Current performance (pure Go):**
- Already delivers 2.91x speedup: 22m59s → 7m54s on M2 Max
- Pure Go loop is highly optimized by Go compiler
- No change with Go 1.26 (ARM64 intrinsics not yet available)

**Future with Go 1.27 NEON intrinsics:**
- Expected additional: +20-40% improvement
- Target: ~5m30s - 6m20s on M2 Max

## How to Verify

### Step 1: Run Correctness Tests

Ensure the SIMD implementation produces identical results to the reference:

```bash
# On any architecture
GOEXPERIMENT=simd go test ./internal/hitable -run TestRayAABB4_SIMD_Correctness -v

# Should see:
# ✓ All hits
# ✓ No hits - ray pointing away
# ✓ tMax cutoff - first two within range
# ✓ Edge case - ray origin inside AABB
# ✓ Negative direction components
# ✓ Mixed directions
# ✓ Infinite ray direction (parallel to axis)
```

### Step 2: Run Random Tests

Test against 1000 random cases:

```bash
GOEXPERIMENT=simd go test ./internal/hitable -run TestRayAABB4_SIMD_RandomCases -v

# Should see:
# ✓ All 1000 random test cases passed
```

### Step 3: Benchmark Comparison

Compare SIMD vs pure Go reference:

```bash
# Benchmark the implementations
GOEXPERIMENT=simd go test ./internal/hitable \
  -bench "BenchmarkRayAABB4_(SIMD|Reference)" \
  -benchmem -benchtime=5s

# Example output on AMD64:
# BenchmarkRayAABB4_SIMD      100000000    8.2 ns/op    0 B/op   0 allocs/op
# BenchmarkRayAABB4_Reference  80000000   10.5 ns/op    0 B/op   0 allocs/op
# => ~22% improvement with intrinsics
```

### Step 4: Full Scene Benchmark

Test with the dragon scene (871,414 triangles):

```bash
cd cmd/izpi

# Run with SIMD (new)
time GOEXPERIMENT=simd ./izpi \
  --x 1920 --y 1080 \
  --scene ../../scenes/dragon.pb.gz \
  --samples 100

# Run without SIMD (old behavior)
time ./izpi \
  --x 1920 --y 1080 \
  --scene ../../scenes/dragon.pb.gz \
  --samples 100

# Compare render times
```

## Build Instructions

### Standard Build (with SIMD)

```bash
# Build the izpi command with SIMD intrinsics
cd cmd/izpi
GOEXPERIMENT=simd go build

# Or use the Makefile
make build
```

### Cross-Compilation

```bash
# Build AMD64 binary on ARM64 (e.g., M2 Mac)
GOEXPERIMENT=simd GOARCH=amd64 go build -o izpi-amd64 ./cmd/izpi

# Build ARM64 binary on AMD64
GOEXPERIMENT=simd GOARCH=arm64 go build -o izpi-arm64 ./cmd/izpi
```

### Building for Distribution

```bash
# Build for multiple architectures
./scripts/build_all.sh  # If you create this script
```

## Code Differences

### Assembly (old) vs Intrinsics (new)

#### Assembly (bvh4_avx2.s) - DELETED ❌

```asm
TEXT ·RayAABB4_SIMD(SB), NOSPLIT, $0-105
    MOVQ    rayOrgX+0(FP), AX
    MOVQ    rayOrgY+8(FP), BX
    VBROADCASTSS (AX), Y0
    VBROADCASTSS (BX), Y1
    VMOVUPS (R8), X6
    VSUBPS  Y0, Y6, Y12
    VMULPS  Y3, Y12, Y12
    // ... 140+ lines of assembly
    VZEROUPPER
    RET
```

**Problems:**
- Hard to read and maintain
- Not inlineable (call overhead)
- Manual register allocation
- No type safety
- Compiler can't optimize across call boundary

#### Go Intrinsics (bvh4_simd_amd64.go) - NEW ✅

```go
//go:inline
func RayAABB4_SIMD(...) uint8 {
    // Broadcast ray parameters
    orgX := archsimd.BroadcastFloat32x8(*rayOrgX)
    invDirX := archsimd.BroadcastFloat32x8(*rayInvDirX)
    
    // Load AABB bounds
    minXVec := archsimd.LoadFloat32x8(&minXArray)
    
    // Compute intersections
    t0x := minXVec.Sub(orgX).Mul(invDirX)
    tNearX := t0x.Min(t1x)
    
    // Extract result
    return finalMask.ToBits() & 0x0F
}
```

**Benefits:**
- Clean, readable Go code
- Type-safe
- Inlineable (eliminates call overhead)
- Compiler can optimize
- Easier to maintain and modify

## API Mapping

Assembly instruction → Go intrinsics method:

| Assembly | Go Intrinsics | Description |
|----------|---------------|-------------|
| `VBROADCASTSS (AX), Y0` | `BroadcastFloat32x8(x)` | Replicate scalar to all lanes |
| `VMOVUPS (R8), X6` | `LoadFloat32x8(&arr)` | Load 8 float32 from memory |
| `VSUBPS Y0, Y6, Y12` | `vec.Sub(other)` | Subtract vectors |
| `VMULPS Y3, Y12, Y12` | `vec.Mul(other)` | Multiply vectors |
| `VMINPS Y12, Y13, Y14` | `vec.Min(other)` | Minimum of vectors |
| `VMAXPS Y12, Y13, Y15` | `vec.Max(other)` | Maximum of vectors |
| `VCMPPS $0x0D, Y0, Y1, Y2` | `vec.GreaterEqual(other)` | Compare, return mask |
| `VANDPS Y2, Y3, Y2` | `mask.And(other)` | AND masks together |
| `VMOVMSKPS Y2, AX` | `mask.ToBits()` | Extract mask to uint8 |
| `VZEROUPPER` | (automatic) | Compiler handles cleanup |

## Troubleshooting

### "could not import simd/archsimd"

Make sure you're building with `GOEXPERIMENT=simd`:

```bash
GOEXPERIMENT=simd go build ./internal/hitable
```

### Cross-compilation on ARM64

When building AMD64 code on ARM64 (like M2 Mac):

```bash
GOEXPERIMENT=simd GOARCH=amd64 go build ...
```

### Performance regression

If you see performance regression:
1. Verify `GOEXPERIMENT=simd` is set
2. Check CPU supports AVX2: `grep avx2 /proc/cpuinfo` (Linux) or `sysctl hw.optional.avx2_0` (macOS)
3. Compare with/without intrinsics using benchmarks
4. Check if CPU governor is in performance mode

## Future: ARM64 NEON Intrinsics (Go 1.27)

When ARM64 support lands in `simd/archsimd`, update `bvh4_simd_arm64.go`:

```go
//go:build arm64

import "simd/archsimd"

func RayAABB4_SIMD(...) uint8 {
    // Use Float32x4 for NEON (128-bit Q registers)
    orgX := archsimd.BroadcastFloat32x4(*rayOrgX)
    
    // Load AABB bounds (4 float32 each)
    minXVec := archsimd.LoadFloat32x4(minX)
    
    // Compute intersections (same algorithm)
    t0x := minXVec.Sub(orgX).Mul(invDirX)
    tNearX := t0x.Min(t1x)
    
    // Extract 4-bit mask
    return finalMask.ToBits()
}
```

**Expected improvements:**
- Eliminate loop overhead
- Reduce branch mispredictions
- Native NEON SIMD parallelism
- ~20-40% faster than pure Go

## References

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [simd/archsimd Package](https://pkg.go.dev/simd/archsimd)
- [GitHub Issue #73787](https://go.dev/issue/73787) - SIMD intrinsics proposal
- [Go SIMD Tutorial](https://callistaenterprise.se/blogg/teknik/2026/01/08/coding-go-for-simd/)

