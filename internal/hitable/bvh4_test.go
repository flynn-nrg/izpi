package hitable

import (
	"fmt"
	"math"
	"testing"

	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/google/go-cmp/cmp"
)

func TestNewBVH4(t *testing.T) {
	testData := []struct {
		name     string
		hitables []Hitable
		time0    float64
		time1    float64
	}{
		{
			name:     "A single sphere",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0)},
			time0:    0,
			time1:    1,
		},
		{
			name:     "Two spheres",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0), makeSphere(1, 0, 0, 1.0)},
			time0:    0,
			time1:    1,
		},
		{
			name:     "Five spheres",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0), makeSphere(1, 0, 0, 1.0), makeSphere(0, 1, 0, 1.0), makeSphere(1, 1, 0, 1.0), makeSphere(1, 1, 1, 1.0)},
			time0:    0,
			time1:    1,
		},
		{
			name: "Ten spheres",
			hitables: []Hitable{
				makeSphere(0, 0, 0, 1.0),
				makeSphere(1, 0, 0, 1.0),
				makeSphere(0, 1, 0, 1.0),
				makeSphere(1, 1, 0, 1.0),
				makeSphere(1, 1, 1, 1.0),
				makeSphere(2, 0, 0, 1.0),
				makeSphere(0, 2, 0, 1.0),
				makeSphere(2, 2, 0, 1.0),
				makeSphere(0, 0, 2, 1.0),
				makeSphere(2, 2, 2, 1.0),
			},
			time0: 0,
			time1: 1,
		},
	}

	ramdonFunc := func() float64 { return 0 }
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := newBVH4(test.hitables, ramdonFunc, test.time0, test.time1)

			// Basic sanity checks
			if got == nil {
				t.Errorf("NewBVH4() returned nil")
				return
			}

			if len(got.Nodes) == 0 {
				t.Errorf("NewBVH4() created BVH with no nodes")
				return
			}

			if len(got.Primitives) != len(test.hitables) {
				t.Errorf("NewBVH4() primitive count = %v, want %v", len(got.Primitives), len(test.hitables))
			}

			// Verify that bounding box is valid
			if box, ok := got.BoundingBox(test.time0, test.time1); !ok || box == nil {
				t.Errorf("NewBVH4() has invalid bounding box")
			}
		})
	}
}

// TestBVH4HitComparison tests that BVH4 produces the same results as binary BVH
func TestBVH4HitComparison(t *testing.T) {
	// Create a scene with multiple spheres
	spheres := []Hitable{
		makeSphere(0, 0, -5, 1.0),
		makeSphere(2, 0, -5, 1.0),
		makeSphere(-2, 0, -5, 1.0),
		makeSphere(0, 2, -5, 1.0),
		makeSphere(0, -2, -5, 1.0),
		makeSphere(1, 1, -5, 0.5),
		makeSphere(-1, -1, -5, 0.5),
	}

	// Build both BVH types with the same random function for consistency
	randomFunc := func() float64 { return 0 }
	bvh2 := newBVH(spheres, randomFunc, 0, 1)
	bvh4 := newBVH4(spheres, randomFunc, 0, 1)

	// Test rays that should hit
	testRays := []struct {
		name string
		ray  ray.Ray
	}{
		{
			name: "Center sphere",
			ray:  ray.New(vec3.Vec3Impl{X: 0, Y: 0, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		},
		{
			name: "Right sphere",
			ray:  ray.New(vec3.Vec3Impl{X: 2, Y: 0, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		},
		{
			name: "Left sphere",
			ray:  ray.New(vec3.Vec3Impl{X: -2, Y: 0, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		},
		{
			name: "Miss all spheres",
			ray:  ray.New(vec3.Vec3Impl{X: 10, Y: 10, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		},
	}

	for _, test := range testRays {
		t.Run(test.name, func(t *testing.T) {
			rec2, mat2, hit2 := bvh2.Hit(test.ray, 0.001, 1000)
			rec4, mat4, hit4 := bvh4.Hit(test.ray, 0.001, 1000)

			// Check if both hit or both miss
			if hit2 != hit4 {
				t.Errorf("Hit mismatch: BVH2 hit=%v, BVH4 hit=%v", hit2, hit4)
				return
			}

			// If both hit, check that hit records are similar
			if hit2 && hit4 {
				if rec2 == nil || rec4 == nil {
					t.Errorf("Hit record is nil: BVH2=%v, BVH4=%v", rec2, rec4)
					return
				}

				// Compare hit distance (T value) with tolerance
				const tolerance = 1e-6
				if diff := rec2.T() - rec4.T(); diff > tolerance || diff < -tolerance {
					t.Errorf("Hit distance mismatch: BVH2 T=%v, BVH4 T=%v", rec2.T(), rec4.T())
				}

				// Verify materials are the same type
				if mat2 == nil || mat4 == nil {
					t.Errorf("Material is nil: BVH2=%v, BVH4=%v", mat2, mat4)
				}
			}
		})
	}
}

// TestRayAABB4_SIMD tests the SIMD AABB intersection function
func TestRayAABB4_SIMD(t *testing.T) {
	// Ray shooting down the Z axis
	// Ray direction: (0, 0, -1)
	// Ray inverse direction: (1/0, 1/0, 1/-1) = (inf, inf, -1)
	rayOrgX := float32(0.0)
	rayOrgY := float32(0.0)
	rayOrgZ := float32(0.0)
	rayInvDirX := float32(math.Inf(1)) // +inf
	rayInvDirY := float32(math.Inf(1)) // +inf
	rayInvDirZ := float32(-1.0)
	tMax := float32(100.0)

	// Four AABBs at different positions
	minX := [4]float32{-1, 10, -1, 10} // Box 0 and 2 should be hit
	minY := [4]float32{-1, -1, 10, 10}
	minZ := [4]float32{-10, -10, -10, -10}
	maxX := [4]float32{1, 12, 1, 12}
	maxY := [4]float32{1, 1, 12, 12}
	maxZ := [4]float32{-2, -2, -2, -2}

	mask := RayAABB4_SIMD(
		&rayOrgX, &rayOrgY, &rayOrgZ,
		&rayInvDirX, &rayInvDirY, &rayInvDirZ,
		&minX, &minY, &minZ,
		&maxX, &maxY, &maxZ,
		tMax,
	)

	// Box 0 should be hit (centered at origin)
	if mask&1 == 0 {
		t.Errorf("Expected box 0 to be hit, but mask = %08b", mask)
	}

	// Box 1 should not be hit (offset in X)
	if mask&2 != 0 {
		t.Errorf("Expected box 1 to not be hit, but mask = %08b", mask)
	}

	// Box 2 should not be hit (offset in Y)
	if mask&4 != 0 {
		t.Errorf("Expected box 2 to not be hit, but mask = %08b", mask)
	}

	// Box 3 should not be hit (offset in X and Y)
	if mask&8 != 0 {
		t.Errorf("Expected box 3 to not be hit, but mask = %08b", mask)
	}
}

// TestRayAABB4_SIMD_AllHit tests when all 4 AABBs should be hit
func TestRayAABB4_SIMD_AllHit(t *testing.T) {
	// Ray shooting down the Z axis
	// Ray direction: (0, 0, -1)
	// Ray inverse direction: (1/0, 1/0, 1/-1) = (inf, inf, -1)
	rayOrgX := float32(0.0)
	rayOrgY := float32(0.0)
	rayOrgZ := float32(0.0)
	rayInvDirX := float32(math.Inf(1)) // +inf
	rayInvDirY := float32(math.Inf(1)) // +inf
	rayInvDirZ := float32(-1.0)
	tMax := float32(100.0)

	// Four AABBs all containing the ray origin
	minX := [4]float32{-2, -2, -2, -2}
	minY := [4]float32{-2, -2, -2, -2}
	minZ := [4]float32{-10, -10, -10, -10}
	maxX := [4]float32{2, 2, 2, 2}
	maxY := [4]float32{2, 2, 2, 2}
	maxZ := [4]float32{-1, -1, -1, -1}

	mask := RayAABB4_SIMD(
		&rayOrgX, &rayOrgY, &rayOrgZ,
		&rayInvDirX, &rayInvDirY, &rayInvDirZ,
		&minX, &minY, &minZ,
		&maxX, &maxY, &maxZ,
		tMax,
	)

	// All boxes should be hit
	expectedMask := uint8(0b1111)
	if mask != expectedMask {
		t.Errorf("Expected all boxes to be hit (mask = %08b), got mask = %08b", expectedMask, mask)
	}
}

// TestRayAABB4_SIMD_NoneHit tests when no AABBs should be hit
func TestRayAABB4_SIMD_NoneHit(t *testing.T) {
	// Ray shooting down the Z axis
	// Ray direction: (0, 0, -1)
	// Ray inverse direction: (1/0, 1/0, 1/-1) = (inf, inf, -1)
	rayOrgX := float32(0.0)
	rayOrgY := float32(0.0)
	rayOrgZ := float32(0.0)
	rayInvDirX := float32(math.Inf(1)) // +inf
	rayInvDirY := float32(math.Inf(1)) // +inf
	rayInvDirZ := float32(-1.0)
	tMax := float32(100.0)

	// Four AABBs all behind the ray
	minX := [4]float32{-2, -2, -2, -2}
	minY := [4]float32{-2, -2, -2, -2}
	minZ := [4]float32{1, 1, 1, 1}
	maxX := [4]float32{2, 2, 2, 2}
	maxY := [4]float32{2, 2, 2, 2}
	maxZ := [4]float32{10, 10, 10, 10}

	mask := RayAABB4_SIMD(
		&rayOrgX, &rayOrgY, &rayOrgZ,
		&rayInvDirX, &rayInvDirY, &rayInvDirZ,
		&minX, &minY, &minZ,
		&maxX, &maxY, &maxZ,
		tMax,
	)

	// No boxes should be hit
	if mask != 0 {
		t.Errorf("Expected no boxes to be hit (mask = 0), got mask = %08b", mask)
	}
}

// TestBVH4BoundingBox tests that the BVH4 bounding box encompasses all primitives
func TestBVH4BoundingBox(t *testing.T) {
	spheres := []Hitable{
		makeSphere(-5, 0, 0, 1.0),
		makeSphere(5, 0, 0, 1.0),
		makeSphere(0, -5, 0, 1.0),
		makeSphere(0, 5, 0, 1.0),
		makeSphere(0, 0, -5, 1.0),
		makeSphere(0, 0, 5, 1.0),
	}

	randomFunc := func() float64 { return 0 }
	bvh4 := newBVH4(spheres, randomFunc, 0, 1)

	box, ok := bvh4.BoundingBox(0, 1)
	if !ok {
		t.Fatalf("BVH4.BoundingBox() returned false")
	}

	// The bounding box should contain all spheres
	// Each sphere has center at ±5 with radius 1, so bounds should be approximately [-6, 6]
	if box.Min().X > -6 || box.Max().X < 6 {
		t.Errorf("BVH4 bounding box X range [%v, %v] doesn't contain all spheres", box.Min().X, box.Max().X)
	}
	if box.Min().Y > -6 || box.Max().Y < 6 {
		t.Errorf("BVH4 bounding box Y range [%v, %v] doesn't contain all spheres", box.Min().Y, box.Max().Y)
	}
	if box.Min().Z > -6 || box.Max().Z < 6 {
		t.Errorf("BVH4 bounding box Z range [%v, %v] doesn't contain all spheres", box.Min().Z, box.Max().Z)
	}
}

// TestBVH4InterfaceMethods tests the other interface methods
func TestBVH4InterfaceMethods(t *testing.T) {
	spheres := []Hitable{makeSphere(0, 0, 0, 1.0)}
	randomFunc := func() float64 { return 0 }
	bvh4 := newBVH4(spheres, randomFunc, 0, 1)

	// Test PDFValue
	pdfVal := bvh4.PDFValue(vec3.Vec3Impl{}, vec3.Vec3Impl{})
	if pdfVal != 0.0 {
		t.Errorf("PDFValue() = %v, want 0.0", pdfVal)
	}

	// Test Random
	random := bvh4.Random(vec3.Vec3Impl{}, nil)
	if diff := cmp.Diff(vec3.Vec3Impl{X: 1}, random); diff != "" {
		t.Errorf("Random() mismatch (-want +got):\n%s", diff)
	}

	// Test IsEmitter
	if bvh4.IsEmitter() {
		t.Errorf("IsEmitter() = true, want false")
	}
}

// TestConservativeFloat32Conversion tests that float64->float32 conversion is conservative
func TestConservativeFloat32Conversion(t *testing.T) {
	// Test values that would lose precision when converting to float32
	testCases := []struct {
		name        string
		value       float64
		testMin     bool // if true, test conservativeFloat32Min, else conservativeFloat32Max
		description string
	}{
		{
			name:        "Large positive value min",
			value:       1234567890.123456789,
			testMin:     true,
			description: "Min conversion should round down or equal",
		},
		{
			name:        "Large positive value max",
			value:       1234567890.123456789,
			testMin:     false,
			description: "Max conversion should round up or equal",
		},
		{
			name:        "Large negative value min",
			value:       -1234567890.123456789,
			testMin:     true,
			description: "Min conversion should round down",
		},
		{
			name:        "Large negative value max",
			value:       -1234567890.123456789,
			testMin:     false,
			description: "Max conversion should round up or equal",
		},
		{
			name:        "Small positive value min",
			value:       0.00000012345678901234,
			testMin:     true,
			description: "Small min conversion should round down or equal",
		},
		{
			name:        "Small positive value max",
			value:       0.00000012345678901234,
			testMin:     false,
			description: "Small max conversion should round up or equal",
		},
		{
			name:        "Value requiring precision min",
			value:       math.Pi * 1000000,
			testMin:     true,
			description: "Pi min conversion should be conservative",
		},
		{
			name:        "Value requiring precision max",
			value:       math.Pi * 1000000,
			testMin:     false,
			description: "Pi max conversion should be conservative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var converted float32
			if tc.testMin {
				converted = conservativeFloat32Min(tc.value)
				// The converted value should be <= original
				if float64(converted) > tc.value {
					t.Errorf("conservativeFloat32Min(%v) = %v (as float64: %v) is greater than original: %s",
						tc.value, converted, float64(converted), tc.description)
				}
			} else {
				converted = conservativeFloat32Max(tc.value)
				// The converted value should be >= original
				if float64(converted) < tc.value {
					t.Errorf("conservativeFloat32Max(%v) = %v (as float64: %v) is less than original: %s",
						tc.value, converted, float64(converted), tc.description)
				}
			}
		})
	}
}

// TestBVH4PrecisionCorrectness tests that BVH4 doesn't miss intersections due to precision
func TestBVH4PrecisionCorrectness(t *testing.T) {
	// Create spheres with float64 positions that will lose precision in float32
	spheres := []Hitable{
		makeSphere(1234567.890123, 0, -10, 1.0),               // Large X coordinate
		makeSphere(0, 0.000001234567, -10, 0.5),               // Small Y coordinate
		makeSphere(-9876543.210987, 100, -10, 2.0),            // Large negative X
		makeSphere(math.Pi*1000000, math.E*1000000, -10, 1.5), // Transcendental numbers
	}

	randomFunc := func() float64 { return 0 }
	bvh2 := newBVH(spheres, randomFunc, 0, 1)
	bvh4 := newBVH4(spheres, randomFunc, 0, 1)

	// Test rays aimed at each sphere
	testRays := []ray.Ray{
		ray.New(vec3.Vec3Impl{X: 1234567.890123, Y: 0, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		ray.New(vec3.Vec3Impl{X: 0, Y: 0.000001234567, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		ray.New(vec3.Vec3Impl{X: -9876543.210987, Y: 100, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		ray.New(vec3.Vec3Impl{X: math.Pi * 1000000, Y: math.E * 1000000, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
	}

	for i, testRay := range testRays {
		t.Run(fmt.Sprintf("Ray_%d", i), func(t *testing.T) {
			_, _, hit2 := bvh2.Hit(testRay, 0.001, 1000)
			_, _, hit4 := bvh4.Hit(testRay, 0.001, 1000)

			// BVH4 should hit everything that BVH2 hits (it can be more conservative)
			if hit2 && !hit4 {
				t.Errorf("BVH4 missed an intersection that BVH2 found at sphere %d. "+
					"This indicates precision loss in float32 conversion.", i)
			}
		})
	}
}

// TestBVH4LargeScene tests BVH4 with many primitives similar to dragon scene
func TestBVH4LargeScene(t *testing.T) {
	// Create a scene with thousands of spheres
	const numSpheres = 10000
	spheres := make([]Hitable, numSpheres)
	for i := 0; i < numSpheres; i++ {
		x := float64(i%100) * 0.5
		y := float64((i/100)%100) * 0.5
		z := float64(i/10000) * 0.5
		spheres[i] = makeSphere(x, y, z-50, 0.2)
	}

	randomFunc := func() float64 { return 0 }
	bvh2 := newBVH(spheres, randomFunc, 0, 1)
	bvh4 := newBVH4(spheres, randomFunc, 0, 1)

	// CRITICAL: Validate primitive count
	if len(bvh4.Primitives) != numSpheres {
		t.Fatalf("BVH4 has %d primitives, expected %d", len(bvh4.Primitives), numSpheres)
	}

	// Validate the BVH4 structure
	if errors := bvh4.Validate(); len(errors) > 0 {
		t.Errorf("BVH4 validation failed with %d errors:", len(errors))
		for i, err := range errors {
			t.Errorf("  Error %d: %s", i+1, err)
			if i >= 4 {
				break
			}
		}
	}

	// Check stats
	stats := bvh4.DebugStats()
	t.Logf("BVH4 Stats: nodes=%v, primitives=%v, leaf_nodes=%v, total_leaf_prims=%v",
		stats["num_nodes"], stats["num_primitives"], stats["leaf_nodes"], stats["total_leaf_primitives"])

	if stats["num_primitives"] != numSpheres {
		t.Errorf("BVH4 stats show %v primitives, expected %d", stats["num_primitives"], numSpheres)
	}

	if stats["total_leaf_primitives"] != numSpheres {
		t.Errorf("BVH4 leaf nodes reference %v primitives, expected %d", stats["total_leaf_primitives"], numSpheres)
	}

	// Test multiple rays
	testRays := []ray.Ray{
		ray.New(vec3.Vec3Impl{X: 5, Y: 5, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		ray.New(vec3.Vec3Impl{X: 10, Y: 10, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
		ray.New(vec3.Vec3Impl{X: 25, Y: 25, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0),
	}

	for i, testRay := range testRays {
		_, _, hit2 := bvh2.Hit(testRay, 0.001, 1000)
		_, _, hit4 := bvh4.Hit(testRay, 0.001, 1000)

		if hit2 != hit4 {
			t.Errorf("Ray %d: BVH4 hit=%v, BVH2 hit=%v (mismatch!)", i, hit4, hit2)
			
			// Additional debug info
			hitRoot := bvh4.TestRayAgainstRoot(testRay, 0.001, 1000)
			t.Errorf("  Ray hits root: %v", hitRoot)
		}
	}
}

// Benchmark the BVH4 Hit method
func BenchmarkBVH4Hit(b *testing.B) {
	// Create a scene with many spheres
	spheres := make([]Hitable, 100)
	for i := 0; i < 100; i++ {
		x := float64(i%10) * 2.0
		y := float64((i/10)%10) * 2.0
		z := float64(i/100) * 2.0
		spheres[i] = makeSphere(x, y, z-50, 0.5)
	}

	randomFunc := func() float64 { return 0 }
	bvh := newBVH4(spheres, randomFunc, 0, 1)

	testRay := ray.New(vec3.Vec3Impl{X: 5, Y: 5, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = bvh.Hit(testRay, 0.001, 1000)
	}
}

// Benchmark comparison between BVH2 and BVH4
func BenchmarkBVHComparison(b *testing.B) {
	// Create a scene with many spheres
	spheres := make([]Hitable, 100)
	for i := 0; i < 100; i++ {
		x := float64(i%10) * 2.0
		y := float64((i/10)%10) * 2.0
		z := float64(i/100) * 2.0
		spheres[i] = makeSphere(x, y, z-50, 0.5)
	}

	randomFunc := func() float64 { return 0 }
	testRay := ray.New(vec3.Vec3Impl{X: 5, Y: 5, Z: 0}, vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, 0)

	b.Run("BVH2", func(b *testing.B) {
		bvh := newBVH(spheres, randomFunc, 0, 1)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = bvh.Hit(testRay, 0.001, 1000)
		}
	})

	b.Run("BVH4", func(b *testing.B) {
		bvh := newBVH4(spheres, randomFunc, 0, 1)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = bvh.Hit(testRay, 0.001, 1000)
		}
	})
}
