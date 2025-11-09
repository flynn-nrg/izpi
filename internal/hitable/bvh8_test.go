package hitable

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Note: makeSphere is defined in bvh_node_test.go and available here

func TestNewBVH8(t *testing.T) {
	spheres := []Hitable{
		makeSphere(0, 0, 0, 1.0),
		makeSphere(3, 0, 0, 1.0),
		makeSphere(-3, 0, 0, 1.0),
	}
	bvh := NewBVH8(spheres, 0.0, 1.0)

	if bvh == nil {
		t.Fatal("NewBVH8 returned nil")
	}

	if len(bvh.Nodes) == 0 {
		t.Error("BVH8 has no nodes")
	}

	if len(bvh.Primitives) != len(spheres) {
		t.Errorf("Expected %d primitives, got %d", len(spheres), len(bvh.Primitives))
	}
}

func TestBVH8HitComparison(t *testing.T) {
	spheres := []Hitable{
		makeSphere(0, 0, 0, 1.0),
		makeSphere(3, 0, 0, 1.0),
		makeSphere(-3, 0, 0, 1.0),
	}
	bvh8 := NewBVH8(spheres, 0.0, 1.0)
	bvh2 := NewBVH(spheres, 0.0, 1.0)

	testCases := []struct {
		name      string
		origin    vec3.Vec3Impl
		direction vec3.Vec3Impl
	}{
		{"Center sphere", vec3.Vec3Impl{X: 0, Y: 0, Z: -10}, vec3.Vec3Impl{X: 0, Y: 0, Z: 1}},
		{"Right sphere", vec3.Vec3Impl{X: 0, Y: 0, Z: -10}, vec3.Vec3Impl{X: 3, Y: 0, Z: 10}},
		{"Left sphere", vec3.Vec3Impl{X: 0, Y: 0, Z: -10}, vec3.Vec3Impl{X: -3, Y: 0, Z: 10}},
		{"Miss all", vec3.Vec3Impl{X: 100, Y: 100, Z: -10}, vec3.Vec3Impl{X: 0, Y: 0, Z: 1}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := ray.New(tc.origin, tc.direction, 0)

			rec8, mat8, hit8 := bvh8.Hit(r, 0.001, 1000.0)
			rec2, mat2, hit2 := bvh2.Hit(r, 0.001, 1000.0)

			if hit8 != hit2 {
				t.Errorf("Hit mismatch: BVH8=%v, BVH2=%v", hit8, hit2)
			}

			if hit8 && hit2 {
				if rec8.T() != rec2.T() {
					t.Errorf("Distance mismatch: BVH8=%v, BVH2=%v", rec8.T(), rec2.T())
				}
				if mat8 != mat2 {
					t.Error("Material mismatch between BVH8 and BVH2")
				}
			}
		})
	}
}

func TestBVH8LargeScene(t *testing.T) {
	const numSpheres = 10000
	spheres := make([]Hitable, numSpheres)
	for i := 0; i < numSpheres; i++ {
		x := float64(i%100 - 50)
		y := float64((i/100)%100 - 50)
		z := float64((i / 10000) - 5)
		spheres[i] = makeSphere(x, y, z, 0.4)
	}

	bvh := NewBVH8(spheres, 0.0, 1.0)

	stats := bvh.DebugStats()
	t.Logf("BVH8 Stats: nodes=%v, primitives=%v, leaf_nodes=%v, total_leaf_prims=%v",
		stats["num_nodes"], stats["num_primitives"], stats["leaf_nodes"], stats["total_leaf_primitives"])

	if stats["num_primitives"].(int) != numSpheres {
		t.Errorf("Expected %d primitives, got %d", numSpheres, stats["num_primitives"])
	}

	if errors := bvh.Validate(); len(errors) > 0 {
		t.Errorf("BVH8 validation failed: %v", errors)
	}
}

func TestBVH8BoundingBox(t *testing.T) {
	spheres := []Hitable{
		makeSphere(0, 0, 0, 1.0),
		makeSphere(3, 0, 0, 1.0),
		makeSphere(-3, 0, 0, 1.0),
	}
	bvh := NewBVH8(spheres, 0.0, 1.0)

	box, ok := bvh.BoundingBox(0.0, 1.0)
	if !ok {
		t.Fatal("BVH8 should have a bounding box")
	}

	if box == nil {
		t.Fatal("Bounding box is nil")
	}

	// Basic sanity checks
	if box.Min().X > box.Max().X || box.Min().Y > box.Max().Y || box.Min().Z > box.Max().Z {
		t.Error("Invalid bounding box: min should be less than max")
	}
}

func TestRayAABB8_SIMD(t *testing.T) {
	// Test the 8-way SIMD function
	rayOrgX := float32(0.0)
	rayOrgY := float32(0.0)
	rayOrgZ := float32(0.0)

	rayInvDirX := float32(1.0)
	rayInvDirY := float32(0.0)
	rayInvDirZ := float32(0.0)

	minX := [8]float32{-1, 1, 2, 3, 4, 5, 6, 7}
	minY := [8]float32{-1, -1, -1, -1, -1, -1, -1, -1}
	minZ := [8]float32{-1, -1, -1, -1, -1, -1, -1, -1}

	maxX := [8]float32{1, 2, 3, 4, 5, 6, 7, 8}
	maxY := [8]float32{1, 1, 1, 1, 1, 1, 1, 1}
	maxZ := [8]float32{1, 1, 1, 1, 1, 1, 1, 1}

	tMax := float32(100.0)

	mask := RayAABB8_SIMD(
		&rayOrgX, &rayOrgY, &rayOrgZ,
		&rayInvDirX, &rayInvDirY, &rayInvDirZ,
		&minX, &minY, &minZ,
		&maxX, &maxY, &maxZ,
		tMax,
	)

	// Ray at origin pointing in +X should hit all 8 boxes
	if mask != 0xFF {
		t.Errorf("Expected mask 0xFF (all 8 bits set), got 0x%02X", mask)
	}
}

func BenchmarkBVH8Hit(b *testing.B) {
	spheres := []Hitable{
		makeSphere(0, 0, 0, 1.0),
		makeSphere(3, 0, 0, 1.0),
		makeSphere(-3, 0, 0, 1.0),
	}
	bvh := NewBVH8(spheres, 0.0, 1.0)
	r := ray.New(vec3.Vec3Impl{X: 0, Y: 0, Z: -10}, vec3.Vec3Impl{X: 0, Y: 0, Z: 1}, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bvh.Hit(r, 0.001, 1000.0)
	}
}

func BenchmarkBVH8Construction(b *testing.B) {
	spheres := createLargeSphereSet(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBVH8(spheres, 0.0, 1.0)
	}
}

func createLargeSphereSet(n int) []Hitable {
	spheres := make([]Hitable, n)
	for i := 0; i < n; i++ {
		x := float64(i%10 - 5)
		y := float64((i/10)%10 - 5)
		z := float64((i / 100) - 5)
		spheres[i] = makeSphere(x, y, z, 0.4)
	}
	return spheres
}

