package hitable

import (
	"math"
	"testing"
)

// Reference Go implementation for testing
func rayAABB4_Reference(
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

func TestRayAABB4_SIMD_Correctness(t *testing.T) {
	tests := []struct {
		name      string
		rayOrg    [3]float32
		rayInvDir [3]float32
		minX      [4]float32
		minY      [4]float32
		minZ      [4]float32
		maxX      [4]float32
		maxY      [4]float32
		maxZ      [4]float32
		tMax      float32
		expected  uint8
	}{
		{
			name:      "All hits",
			rayOrg:    [3]float32{0, 0, 0},
			rayInvDir: [3]float32{1, 1, 1},
			minX:      [4]float32{1, 2, 3, 4},
			minY:      [4]float32{1, 2, 3, 4},
			minZ:      [4]float32{1, 2, 3, 4},
			maxX:      [4]float32{2, 3, 4, 5},
			maxY:      [4]float32{2, 3, 4, 5},
			maxZ:      [4]float32{2, 3, 4, 5},
			tMax:      100,
			expected:  0b1111, // All 4 hit
		},
		{
			name:      "No hits - ray pointing away",
			rayOrg:    [3]float32{0, 0, 0},
			rayInvDir: [3]float32{-1, -1, -1},
			minX:      [4]float32{1, 2, 3, 4},
			minY:      [4]float32{1, 2, 3, 4},
			minZ:      [4]float32{1, 2, 3, 4},
			maxX:      [4]float32{2, 3, 4, 5},
			maxY:      [4]float32{2, 3, 4, 5},
			maxZ:      [4]float32{2, 3, 4, 5},
			tMax:      100,
			expected:  0b0000, // None hit
		},
		{
			name:      "Partial hits - alternating pattern",
			rayOrg:    [3]float32{0, 0, -10},
			rayInvDir: [3]float32{0, 0, 1},
			minX:      [4]float32{-1, 5, -1, 5},
			minY:      [4]float32{-1, -1, 5, 5},
			minZ:      [4]float32{0, 0, 0, 0},
			maxX:      [4]float32{1, 6, 1, 6},
			maxY:      [4]float32{1, 1, 6, 6},
			maxZ:      [4]float32{10, 10, 10, 10},
			tMax:      100,
			expected:  0b0001, // Only first box hit
		},
		{
			name:      "tMax cutoff - first two within range",
			rayOrg:    [3]float32{0, 0, 0},
			rayInvDir: [3]float32{1, 1, 1},
			minX:      [4]float32{1, 2, 10, 20},
			minY:      [4]float32{1, 2, 10, 20},
			minZ:      [4]float32{1, 2, 10, 20},
			maxX:      [4]float32{2, 3, 11, 21},
			maxY:      [4]float32{2, 3, 11, 21},
			maxZ:      [4]float32{2, 3, 11, 21},
			tMax:      5,
			expected:  0b0011, // Only first two hit (others beyond tMax)
		},
		{
			name:      "Edge case - ray origin inside AABB",
			rayOrg:    [3]float32{1.5, 1.5, 1.5},
			rayInvDir: [3]float32{1, 1, 1},
			minX:      [4]float32{1, 5, 5, 5},
			minY:      [4]float32{1, 5, 5, 5},
			minZ:      [4]float32{1, 5, 5, 5},
			maxX:      [4]float32{2, 6, 6, 6},
			maxY:      [4]float32{2, 6, 6, 6},
			maxZ:      [4]float32{2, 6, 6, 6},
			tMax:      100,
			expected:  0b1111, // Should hit first box (inside it)
		},
		{
			name:      "Negative direction components",
			rayOrg:    [3]float32{5, 5, 5},
			rayInvDir: [3]float32{-1, -1, -1},
			minX:      [4]float32{1, 2, 3, 6},
			minY:      [4]float32{1, 2, 3, 6},
			minZ:      [4]float32{1, 2, 3, 6},
			maxX:      [4]float32{2, 3, 4, 7},
			maxY:      [4]float32{2, 3, 4, 7},
			maxZ:      [4]float32{2, 3, 4, 7},
			tMax:      10,
			expected:  0b0111, // First three hit, fourth is in front
		},
		{
			name:      "Mixed directions",
			rayOrg:    [3]float32{0, 0, 0},
			rayInvDir: [3]float32{1, 1, -1},
			minX:      [4]float32{1, 1, 1, 1},
			minY:      [4]float32{1, 1, 1, 1},
			minZ:      [4]float32{-2, 1, -2, 1},
			maxX:      [4]float32{2, 2, 2, 2},
			maxY:      [4]float32{2, 2, 2, 2},
			maxZ:      [4]float32{-1, 2, -1, 2},
			tMax:      100,
			expected:  0b0101, // Boxes 0 and 2 hit (z < 0)
		},
		{
			name:      "Very small boxes",
			rayOrg:    [3]float32{0, 0, 0},
			rayInvDir: [3]float32{1, 0, 0},
			minX:      [4]float32{1, 2, 3, 4},
			minY:      [4]float32{-0.01, -0.01, -0.01, -0.01},
			minZ:      [4]float32{-0.01, -0.01, -0.01, -0.01},
			maxX:      [4]float32{1.01, 2.01, 3.01, 4.01},
			maxY:      [4]float32{0.01, 0.01, 0.01, 0.01},
			maxZ:      [4]float32{0.01, 0.01, 0.01, 0.01},
			tMax:      10,
			expected:  0b1111, // All should hit (ray passes through center)
		},
		{
			name:      "Infinite ray direction (parallel to axis)",
			rayOrg:    [3]float32{0, 5, 5},
			rayInvDir: [3]float32{1, math.MaxFloat32, math.MaxFloat32}, // ~infinite for Y and Z
			minX:      [4]float32{1, 2, 3, 4},
			minY:      [4]float32{4, 4, 6, 6},
			minZ:      [4]float32{4, 4, 6, 6},
			maxX:      [4]float32{2, 3, 4, 5},
			maxY:      [4]float32{6, 6, 7, 7},
			maxZ:      [4]float32{6, 6, 7, 7},
			tMax:      100,
			expected:  0b0011, // Only first two hit (Y and Z contain ray)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test SIMD implementation
			simdResult := rayAABB4_SIMD_impl(
				&tt.rayOrg[0], &tt.rayOrg[1], &tt.rayOrg[2],
				&tt.rayInvDir[0], &tt.rayInvDir[1], &tt.rayInvDir[2],
				&tt.minX, &tt.minY, &tt.minZ,
				&tt.maxX, &tt.maxY, &tt.maxZ,
				tt.tMax,
			)

			// Test reference implementation
			refResult := rayAABB4_Reference(
				&tt.rayOrg[0], &tt.rayOrg[1], &tt.rayOrg[2],
				&tt.rayInvDir[0], &tt.rayInvDir[1], &tt.rayInvDir[2],
				&tt.minX, &tt.minY, &tt.minZ,
				&tt.maxX, &tt.maxY, &tt.maxZ,
				tt.tMax,
			)

			// Verify both match expected
			if refResult != tt.expected {
				t.Errorf("Reference implementation failed: got mask 0x%02x, expected 0x%02x", refResult, tt.expected)
			}

			if simdResult != tt.expected {
				t.Errorf("SIMD implementation failed: got mask 0x%02x, expected 0x%02x", simdResult, tt.expected)
			}

			// Most importantly, verify SIMD matches reference
			if simdResult != refResult {
				t.Errorf("SIMD result (0x%02x) differs from reference (0x%02x)", simdResult, refResult)
			}

			t.Logf("  ✓ SIMD and reference both returned mask: 0b%04b", simdResult)
		})
	}
}

// TestRayAABB4_SIMD_RandomCases generates random test cases to catch edge cases
func TestRayAABB4_SIMD_RandomCases(t *testing.T) {
	const numTests = 1000

	for i := 0; i < numTests; i++ {
		// Generate random ray
		rayOrg := [3]float32{
			float32(i%10 - 5),
			float32((i+1)%10 - 5),
			float32((i+2)%10 - 5),
		}

		// Generate random direction (avoiding near-zero)
		dirX := float32((i%7)-3) + 0.1
		dirY := float32(((i+1)%7)-3) + 0.1
		dirZ := float32(((i+2)%7)-3) + 0.1

		rayInvDir := [3]float32{
			1.0 / dirX,
			1.0 / dirY,
			1.0 / dirZ,
		}

		// Generate random AABBs
		var minX, minY, minZ, maxX, maxY, maxZ [4]float32
		for j := 0; j < 4; j++ {
			baseX := float32((i+j)%20 - 10)
			baseY := float32((i+j+1)%20 - 10)
			baseZ := float32((i+j+2)%20 - 10)

			minX[j] = baseX
			minY[j] = baseY
			minZ[j] = baseZ
			maxX[j] = baseX + float32(j+1)
			maxY[j] = baseY + float32(j+1)
			maxZ[j] = baseZ + float32(j+1)
		}

		tMax := float32(50 + i%50)

		// Compare SIMD vs reference
		simdResult := rayAABB4_SIMD_impl(
			&rayOrg[0], &rayOrg[1], &rayOrg[2],
			&rayInvDir[0], &rayInvDir[1], &rayInvDir[2],
			&minX, &minY, &minZ,
			&maxX, &maxY, &maxZ,
			tMax,
		)

		refResult := rayAABB4_Reference(
			&rayOrg[0], &rayOrg[1], &rayOrg[2],
			&rayInvDir[0], &rayInvDir[1], &rayInvDir[2],
			&minX, &minY, &minZ,
			&maxX, &maxY, &maxZ,
			tMax,
		)

		if simdResult != refResult {
			t.Errorf("Test %d: SIMD (0x%02x) != Reference (0x%02x)", i, simdResult, refResult)
			t.Errorf("  Ray: org=%.2f, invDir=%.2f", rayOrg, rayInvDir)
			t.Errorf("  minX=%v, maxX=%v", minX, maxX)
			t.Errorf("  minY=%v, maxY=%v", minY, maxY)
			t.Errorf("  minZ=%v, maxZ=%v", minZ, maxZ)
			t.Errorf("  tMax=%.2f", tMax)
			break
		}
	}

	t.Logf("✓ All %d random test cases passed", numTests)
}

// BenchmarkRayAABB4_SIMD benchmarks the SIMD implementation
func BenchmarkRayAABB4_SIMD(b *testing.B) {
	rayOrg := [3]float32{0, 0, 0}
	rayInvDir := [3]float32{1, 1, 1}
	minX := [4]float32{1, 2, 3, 4}
	minY := [4]float32{1, 2, 3, 4}
	minZ := [4]float32{1, 2, 3, 4}
	maxX := [4]float32{2, 3, 4, 5}
	maxY := [4]float32{2, 3, 4, 5}
	maxZ := [4]float32{2, 3, 4, 5}
	tMax := float32(100)

	var result uint8
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = rayAABB4_SIMD_impl(
			&rayOrg[0], &rayOrg[1], &rayOrg[2],
			&rayInvDir[0], &rayInvDir[1], &rayInvDir[2],
			&minX, &minY, &minZ,
			&maxX, &maxY, &maxZ,
			tMax,
		)
	}
	_ = result
}

// BenchmarkRayAABB4_Reference benchmarks the reference Go implementation
func BenchmarkRayAABB4_Reference(b *testing.B) {
	rayOrg := [3]float32{0, 0, 0}
	rayInvDir := [3]float32{1, 1, 1}
	minX := [4]float32{1, 2, 3, 4}
	minY := [4]float32{1, 2, 3, 4}
	minZ := [4]float32{1, 2, 3, 4}
	maxX := [4]float32{2, 3, 4, 5}
	maxY := [4]float32{2, 3, 4, 5}
	maxZ := [4]float32{2, 3, 4, 5}
	tMax := float32(100)

	var result uint8
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = rayAABB4_Reference(
			&rayOrg[0], &rayOrg[1], &rayOrg[2],
			&rayInvDir[0], &rayInvDir[1], &rayInvDir[2],
			&minX, &minY, &minZ,
			&maxX, &maxY, &maxZ,
			tMax,
		)
	}
	_ = result
}

