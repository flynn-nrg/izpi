package hitable

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"

	log "github.com/sirupsen/logrus"
)

// Ensure interface compliance.
var _ Hitable = (*BVH4)(nil)

// A single node in the flat BVH array. All bounds are float32.
type BVH4Node struct {
	// The 6 bounds arrays are for the 4 children (4 * 6 * 4 bytes = 96 bytes)
	MinX [4]float32
	MinY [4]float32
	MinZ [4]float32
	MaxX [4]float32
	MaxY [4]float32
	MaxZ [4]float32

	// ChildIndex: Index into the Nodes array (inner node) or Primitives array (leaf)
	ChildIndex [4]int32

	// PrimitiveCount:
	// > 0 for a leaf (number of objects starting at Primitives[ChildIndex[i]])
	// = 0 for an inner node
	PrimitiveCount [4]int32
}

// BVH4 represents a bounding volume hierarchy.
type BVH4 struct {
	Nodes      []BVH4Node
	Primitives []Hitable
	time0      float64
	time1      float64
}

func (bvh *BVH4) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	// Early exit if no nodes
	if len(bvh.Nodes) == 0 {
		return nil, nil, false
	}

	// Current best hit record found so far.
	var bestHitRec *hitrecord.HitRecord
	var bestMat material.Material
	hitFound := false

	// Ray data conversion: Convert ray (likely float64) to float32 for SIMD
	rayInvDir := vec3.Vec3Impl{
		X: 1.0 / r.Direction().X,
		Y: 1.0 / r.Direction().Y,
		Z: 1.0 / r.Direction().Z,
	}
	rInvX, rInvY, rInvZ := float32(rayInvDir.X), float32(rayInvDir.Y), float32(rayInvDir.Z)
	rOrgX, rOrgY, rOrgZ := float32(r.Origin().X), float32(r.Origin().Y), float32(r.Origin().Z)

	// Traversal Stack Setup (Stack of node indices to visit)
	// A fixed-size array is faster than a dynamic slice for small depths (e.g., 64)
	const MAX_DEPTH = 64
	stack := [MAX_DEPTH]int32{}
	stackPtr := int32(0)

	// Start at the root node (index 0)
	currentNodeIndex := int32(0)

	// Main Traversal Loop
	for {
		if currentNodeIndex == -1 {
			// Stack is empty, traversal finished.
			break
		}

		// Safety check (optional but recommended)
		if currentNodeIndex >= int32(len(bvh.Nodes)) {
			// Should not happen in a correctly built tree
			break
		}

		node := bvh.Nodes[currentNodeIndex]

		// --- 1. SIMD Traversal Call ---
		// Pass pointers to the ray data and the node's SoA bounds arrays.
		mask := RayAABB4_SIMD(
			&rOrgX, &rOrgY, &rOrgZ,
			&rInvX, &rInvY, &rInvZ,
			&node.MinX, &node.MinY, &node.MinZ,
			&node.MaxX, &node.MaxY, &node.MaxZ,
			float32(tMax), // Pass current best hit distance
		)

		// --- 2. Mask Processing and Stack Management ---
		nextChildIndex := int32(-1) // Index of the next node to immediately visit

		// Iterate through the 4 potential children (0 to 3)
		// Note: The order of iteration can be optimized (e.g., front-to-back),
		// but simple 0-to-3 is correct for a stack-based DFS.
		for i := int32(0); i < 4; i++ {
			// Check if the i-th bit in the mask is set
			if (mask>>i)&1 == 0 {
				continue // Child was culled by the SIMD test
			}

			childIndex := node.ChildIndex[i]
			primitiveCount := node.PrimitiveCount[i]

			// Skip invalid children (should not happen if AABB test is correct)
			if childIndex == -1 {
				continue
			}

			if primitiveCount > 0 {
				// --- LEAF NODE: Perform scalar ray-primitive tests ---
				for p := int32(0); p < primitiveCount; p++ {
					primitive := bvh.Primitives[childIndex+p]
					if rec, mat, hit := primitive.Hit(r, tMin, tMax); hit {
						// Update the closest hit and shorten the max search distance
						tMax = rec.T()
						bestHitRec = rec
						bestMat = mat
						hitFound = true
					}
				}
			} else {
				// --- INNER NODE: Decide whether to visit now or push to stack ---
				if nextChildIndex == -1 {
					// Always immediately visit the first found hit child to maintain
					// depth-first traversal and reduce stack pushes.
					nextChildIndex = childIndex
				} else {
					// Push all subsequent hit children onto the stack for later
					stack[stackPtr] = childIndex
					stackPtr++
				}
			}
		}

		// --- 3. Update Current Node and Stack ---
		if nextChildIndex != -1 {
			// Visit the child we chose to process immediately
			currentNodeIndex = nextChildIndex
		} else if stackPtr > 0 {
			// Pop the next node index from the stack (LIFO)
			stackPtr--
			currentNodeIndex = stack[stackPtr]
		} else {
			// Stack empty, exit loop
			currentNodeIndex = -1
		}
	}

	return bestHitRec, bestMat, hitFound
}

func (bvh *BVH4) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
	// Early exit if no nodes
	if len(bvh.Nodes) == 0 {
		return nil, false, false
	}

	// Current best hit record found so far.
	var bestHitRec *hitrecord.HitRecord
	var bestHitEdge bool
	hitFound := false

	// Ray data conversion: Convert ray (likely float64) to float32 for SIMD
	rayInvDir := vec3.Vec3Impl{
		X: 1.0 / r.Direction().X,
		Y: 1.0 / r.Direction().Y,
		Z: 1.0 / r.Direction().Z,
	}
	rInvX, rInvY, rInvZ := float32(rayInvDir.X), float32(rayInvDir.Y), float32(rayInvDir.Z)
	rOrgX, rOrgY, rOrgZ := float32(r.Origin().X), float32(r.Origin().Y), float32(r.Origin().Z)

	// Traversal Stack Setup
	const MAX_DEPTH = 64
	stack := [MAX_DEPTH]int32{}
	stackPtr := int32(0)

	// Start at the root node (index 0)
	currentNodeIndex := int32(0)

	// Main Traversal Loop
	for {
		if currentNodeIndex == -1 {
			break
		}

		if currentNodeIndex >= int32(len(bvh.Nodes)) {
			break
		}

		node := bvh.Nodes[currentNodeIndex]

		// SIMD Traversal Call
		mask := RayAABB4_SIMD(
			&rOrgX, &rOrgY, &rOrgZ,
			&rInvX, &rInvY, &rInvZ,
			&node.MinX, &node.MinY, &node.MinZ,
			&node.MaxX, &node.MaxY, &node.MaxZ,
			float32(tMax),
		)

		nextChildIndex := int32(-1)

		for i := int32(0); i < 4; i++ {
			if (mask>>i)&1 == 0 {
				continue
			}

			childIndex := node.ChildIndex[i]
			primitiveCount := node.PrimitiveCount[i]

			// Skip invalid children (should not happen if AABB test is correct)
			if childIndex == -1 {
				continue
			}

			if primitiveCount > 0 {
				// LEAF NODE: Perform scalar ray-primitive tests
				for p := int32(0); p < primitiveCount; p++ {
					primitive := bvh.Primitives[childIndex+p]
					if rec, hit, hitEdge := primitive.HitEdge(r, tMin, tMax); hit {
						tMax = rec.T()
						bestHitRec = rec
						bestHitEdge = hitEdge
						hitFound = true
					}
				}
			} else {
				// INNER NODE: Decide whether to visit now or push to stack
				if nextChildIndex == -1 {
					nextChildIndex = childIndex
				} else {
					stack[stackPtr] = childIndex
					stackPtr++
				}
			}
		}

		if nextChildIndex != -1 {
			currentNodeIndex = nextChildIndex
		} else if stackPtr > 0 {
			stackPtr--
			currentNodeIndex = stack[stackPtr]
		} else {
			currentNodeIndex = -1
		}
	}

	return bestHitRec, hitFound, bestHitEdge
}

func (bvh *BVH4) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	if len(bvh.Nodes) == 0 {
		return nil, false
	}

	// Return the bounding box of the root node (first node)
	root := bvh.Nodes[0]

	// Find the overall bounds across all 4 children of root
	minX := float64(root.MinX[0])
	minY := float64(root.MinY[0])
	minZ := float64(root.MinZ[0])
	maxX := float64(root.MaxX[0])
	maxY := float64(root.MaxY[0])
	maxZ := float64(root.MaxZ[0])

	for i := 1; i < 4; i++ {
		if root.ChildIndex[i] != -1 {
			minX = math.Min(minX, float64(root.MinX[i]))
			minY = math.Min(minY, float64(root.MinY[i]))
			minZ = math.Min(minZ, float64(root.MinZ[i]))
			maxX = math.Max(maxX, float64(root.MaxX[i]))
			maxY = math.Max(maxY, float64(root.MaxY[i]))
			maxZ = math.Max(maxZ, float64(root.MaxZ[i]))
		}
	}

	return aabb.New(
		vec3.Vec3Impl{X: minX, Y: minY, Z: minZ},
		vec3.Vec3Impl{X: maxX, Y: maxY, Z: maxZ},
	), true
}

func (bvh *BVH4) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float64 {
	return 0.0
}

func (bvh *BVH4) Random(o vec3.Vec3Impl, _ *fastrandom.LCG) vec3.Vec3Impl {
	return vec3.Vec3Impl{X: 1}
}

func (bvh *BVH4) IsEmitter() bool {
	return false
}

// DebugStats returns statistics about the BVH4 structure
func (bvh *BVH4) DebugStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["num_nodes"] = len(bvh.Nodes)
	stats["num_primitives"] = len(bvh.Primitives)

	// Count leaf nodes and inner nodes
	leafNodes := 0
	innerNodes := 0
	totalLeafPrimitives := 0
	emptySlots := 0

	for _, node := range bvh.Nodes {
		hasLeaf := false
		hasInner := false
		for i := 0; i < 4; i++ {
			if node.ChildIndex[i] == -1 {
				emptySlots++
				continue
			}
			if node.PrimitiveCount[i] > 0 {
				hasLeaf = true
				totalLeafPrimitives += int(node.PrimitiveCount[i])
			} else {
				hasInner = true
			}
		}
		if hasLeaf {
			leafNodes++
		}
		if hasInner {
			innerNodes++
		}
	}

	stats["leaf_nodes"] = leafNodes
	stats["inner_nodes"] = innerNodes
	stats["total_leaf_primitives"] = totalLeafPrimitives
	stats["empty_slots"] = emptySlots
	stats["avg_primitives_per_slot"] = float64(len(bvh.Primitives)) / float64(len(bvh.Nodes)*4-emptySlots)

	// Check root node
	if len(bvh.Nodes) > 0 {
		root := bvh.Nodes[0]
		rootChildren := 0
		for i := 0; i < 4; i++ {
			if root.ChildIndex[i] != -1 {
				rootChildren++
			}
		}
		stats["root_children"] = rootChildren

		// Root bounding box
		stats["root_bounds"] = map[string]interface{}{
			"min": []float32{root.MinX[0], root.MinY[0], root.MinZ[0]},
			"max": []float32{root.MaxX[0], root.MaxY[0], root.MaxZ[0]},
		}
	}

	return stats
}

// TestRayAgainstRoot tests if a ray hits the root bounding box
func (bvh *BVH4) TestRayAgainstRoot(r ray.Ray, tMin float64, tMax float64) bool {
	if len(bvh.Nodes) == 0 {
		return false
	}

	root := bvh.Nodes[0]
	rayInvDir := vec3.Vec3Impl{
		X: 1.0 / r.Direction().X,
		Y: 1.0 / r.Direction().Y,
		Z: 1.0 / r.Direction().Z,
	}
	rInvX, rInvY, rInvZ := float32(rayInvDir.X), float32(rayInvDir.Y), float32(rayInvDir.Z)
	rOrgX, rOrgY, rOrgZ := float32(r.Origin().X), float32(r.Origin().Y), float32(r.Origin().Z)

	mask := RayAABB4_SIMD(
		&rOrgX, &rOrgY, &rOrgZ,
		&rInvX, &rInvY, &rInvZ,
		&root.MinX, &root.MinY, &root.MinZ,
		&root.MaxX, &root.MaxY, &root.MaxZ,
		float32(tMax),
	)

	return mask != 0
}

// Validate checks the integrity of the BVH4 structure and returns any errors found
func (bvh *BVH4) Validate() []string {
	var errors []string

	if len(bvh.Nodes) == 0 {
		errors = append(errors, "BVH4 has no nodes")
		return errors
	}

	if len(bvh.Primitives) == 0 {
		errors = append(errors, "BVH4 has no primitives")
		return errors
	}

	// Count all primitives referenced in leaf nodes
	primitivesReferenced := 0
	visited := make(map[int32]bool)
	toVisit := []int32{0} // Start with root

	for len(toVisit) > 0 {
		nodeIdx := toVisit[0]
		toVisit = toVisit[1:]

		if nodeIdx < 0 || nodeIdx >= int32(len(bvh.Nodes)) {
			errors = append(errors, fmt.Sprintf("Invalid node index: %d", nodeIdx))
			continue
		}

		if visited[nodeIdx] {
			continue
		}
		visited[nodeIdx] = true

		node := bvh.Nodes[nodeIdx]

		for i := 0; i < 4; i++ {
			childIdx := node.ChildIndex[i]
			primCount := node.PrimitiveCount[i]

			if childIdx == -1 {
				// Invalid slot, skip
				continue
			}

			if primCount > 0 {
				// Leaf node - check primitive indices are valid
				for p := int32(0); p < primCount; p++ {
					idx := childIdx + p
					if idx < 0 || idx >= int32(len(bvh.Primitives)) {
						errors = append(errors, fmt.Sprintf("Leaf node %d slot %d references invalid primitive index %d (count %d, primitives len %d)",
							nodeIdx, i, idx, primCount, len(bvh.Primitives)))
					} else {
						primitivesReferenced++
					}
				}
			} else {
				// Inner node - add to visit list
				toVisit = append(toVisit, childIdx)
			}
		}
	}

	if primitivesReferenced != len(bvh.Primitives) {
		errors = append(errors, fmt.Sprintf("Primitive count mismatch: %d referenced in leaves, %d in primitives array",
			primitivesReferenced, len(bvh.Primitives)))
	}

	return errors
}

// RayAABB4_SIMD tests a ray against 4 AABBs simultaneously.
// Returns a 4-bit mask where bit i is set if the ray hits AABB i.
// Implemented in platform-specific files with build tags:
// - bvh4_simd_amd64.go - AMD64 AVX2 assembly
// - bvh4_simd_arm64.go - ARM64 pure Go
// - bvh4_simd_generic.go - Other platforms pure Go fallback

// Helper functions for float32 min/max
func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// conservativeFloat32Min converts a float64 to float32, rounding DOWN (towards -infinity).
// This ensures the converted minimum bound is always <= the original value.
func conservativeFloat32Min(val float64) float32 {
	f32 := float32(val)
	// Check if rounding occurred by converting back
	if float64(f32) > val {
		// The conversion rounded up, so we need to use the next lower float32 value
		return math.Nextafter32(f32, float32(math.Inf(-1)))
	}
	return f32
}

// conservativeFloat32Max converts a float64 to float32, rounding UP (towards +infinity).
// This ensures the converted maximum bound is always >= the original value.
func conservativeFloat32Max(val float64) float32 {
	f32 := float32(val)
	// Check if rounding occurred by converting back
	if float64(f32) < val {
		// The conversion rounded down, so we need to use the next higher float32 value
		return math.Nextafter32(f32, float32(math.Inf(1)))
	}
	return f32
}

// NewBVH4 builds a 4-way BVH from a list of hitables.
func NewBVH4(hitables []Hitable, time0 float64, time1 float64) *BVH4 {
	log.Infof("Building BVH4 with %v elements", len(hitables))
	startTime := time.Now()
	randomFunc := fastrandom.NewWithDefaults().Float64
	bvh := newBVH4(hitables, randomFunc, time0, time1)
	log.Infof("Completed BVH4 construction in %v", time.Since(startTime))

	// Log debug statistics
	stats := bvh.DebugStats()
	log.Infof("BVH4 Stats: nodes=%v, primitives=%v, leaf_nodes=%v, inner_nodes=%v, total_leaf_prims=%v, empty_slots=%v, root_children=%v",
		stats["num_nodes"], stats["num_primitives"], stats["leaf_nodes"], stats["inner_nodes"],
		stats["total_leaf_primitives"], stats["empty_slots"], stats["root_children"])

	if rootBounds, ok := stats["root_bounds"].(map[string]interface{}); ok {
		log.Infof("BVH4 Root bounds: min=%v, max=%v", rootBounds["min"], rootBounds["max"])
	}

	// Validate the structure
	if errors := bvh.Validate(); len(errors) > 0 {
		log.Errorf("BVH4 validation failed with %d errors:", len(errors))
		for i, err := range errors {
			log.Errorf("  Error %d: %s", i+1, err)
			if i >= 9 { // Limit to first 10 errors
				log.Errorf("  ... and %d more errors", len(errors)-10)
				break
			}
		}
	} else {
		log.Info("BVH4 validation passed")
	}

	return bvh
}

// buildNode is a helper struct for building the BVH4 tree
type buildNode struct {
	box              *aabb.AABB
	children         []*buildNode
	primitiveIndices []int // Indices into the original primitives array
}

func newBVH4(hitables []Hitable, randomFunc func() float64, time0 float64, time1 float64) *BVH4 {
	if len(hitables) == 0 {
		log.Error("Cannot create BVH4 with no hitables")
		return nil
	}

	bvh := &BVH4{
		time0:      time0,
		time1:      time1,
		Primitives: hitables,
	}

	// Create initial index array [0, 1, 2, ..., n-1]
	indices := make([]int, len(hitables))
	for i := range indices {
		indices[i] = i
	}

	// Build a traditional binary BVH first, tracking indices
	root := buildBinaryBVH(hitables, indices, randomFunc, time0, time1)

	// Convert to 4-way BVH by flattening and collapsing levels
	bvh.Nodes = make([]BVH4Node, 0)
	primitiveIndices := make([]int, 0)

	flattenBVH4(root, bvh, &primitiveIndices)

	// Reorder primitives based on the flattening
	reorderedPrimitives := make([]Hitable, len(primitiveIndices))
	for i, idx := range primitiveIndices {
		reorderedPrimitives[i] = hitables[idx]
	}
	bvh.Primitives = reorderedPrimitives

	return bvh
}

// buildBinaryBVH builds a traditional binary BVH tree
func buildBinaryBVH(hitables []Hitable, indices []int, randomFunc func() float64, time0 float64, time1 float64) *buildNode {
	if len(hitables) == 0 {
		return nil
	}

	node := &buildNode{}

	// Compute bounding box for all primitives
	if len(hitables) == 1 {
		node.primitiveIndices = indices
		if box, ok := hitables[0].BoundingBox(time0, time1); ok {
			node.box = box
		}
		return node
	}

	// Compute overall bounding box
	var overallBox *aabb.AABB
	for _, h := range hitables {
		if box, ok := h.BoundingBox(time0, time1); ok {
			if overallBox == nil {
				overallBox = box
			} else {
				overallBox = aabb.SurroundingBox(overallBox, box)
			}
		}
	}
	node.box = overallBox

	// Select a random axis for splitting (0=x, 1=y, 2=z)
	axis := int(3 * randomFunc())

	// Make working copies to avoid corrupting shared slice memory
	// This MUST be done BEFORE sorting!
	workingHitables := make([]Hitable, len(hitables))
	workingIndices := make([]int, len(indices))
	copy(workingHitables, hitables)
	copy(workingIndices, indices)

	// Sort both hitables and indices together along the chosen axis
	sortHitablesWithIndices(workingHitables, workingIndices, axis, time0, time1)

	if len(workingHitables) <= 4 {
		// Create leaf node
		node.primitiveIndices = workingIndices
		return node
	}

	// Split in the middle
	mid := len(workingHitables) / 2

	leftChild := buildBinaryBVH(workingHitables[:mid], workingIndices[:mid], randomFunc, time0, time1)
	rightChild := buildBinaryBVH(workingHitables[mid:], workingIndices[mid:], randomFunc, time0, time1)

	node.children = []*buildNode{leftChild, rightChild}
	return node
}

// sortHitablesWithIndices sorts hitables and their indices together along the specified axis
func sortHitablesWithIndices(hitables []Hitable, indices []int, axis int, time0 float64, time1 float64) {
	// Wrapper to sort both slices together
	type pair struct {
		hitable Hitable
		index   int
	}

	pairs := make([]pair, len(hitables))
	for i := range hitables {
		pairs[i] = pair{hitables[i], indices[i]}
	}

	switch axis {
	case 0: // X
		sort.Slice(pairs, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = pairs[i].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = pairs[j].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessX(box0, box1)
		})
	case 1: // Y
		sort.Slice(pairs, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = pairs[i].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = pairs[j].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessY(box0, box1)
		})
	case 2: // Z
		sort.Slice(pairs, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = pairs[i].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = pairs[j].hitable.BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessZ(box0, box1)
		})
	}

	// Copy sorted pairs back
	for i := range pairs {
		hitables[i] = pairs[i].hitable
		indices[i] = pairs[i].index
	}
}

// flattenBVH4 converts a binary BVH tree into a flat 4-way BVH structure
func flattenBVH4(node *buildNode, bvh *BVH4, primitiveIndices *[]int) int32 {
	if node == nil {
		return -1
	}

	nodeIndex := int32(len(bvh.Nodes))
	bvh4Node := BVH4Node{}

	// Initialize all children as invalid
	// Use a degenerate AABB that will always fail the ray test
	// (set all bounds to MaxFloat32 so tNear will be huge and fail tNear <= tMax)
	for i := 0; i < 4; i++ {
		bvh4Node.ChildIndex[i] = -1
		bvh4Node.PrimitiveCount[i] = 0
		bvh4Node.MinX[i] = math.MaxFloat32
		bvh4Node.MinY[i] = math.MaxFloat32
		bvh4Node.MinZ[i] = math.MaxFloat32
		bvh4Node.MaxX[i] = math.MaxFloat32
		bvh4Node.MaxY[i] = math.MaxFloat32
		bvh4Node.MaxZ[i] = math.MaxFloat32
	}

	// Check if this is a leaf node
	if len(node.primitiveIndices) > 0 {
		// This is a leaf - store primitives
		primStart := int32(len(*primitiveIndices))

		// Add primitive indices
		*primitiveIndices = append(*primitiveIndices, node.primitiveIndices...)

		// Store in first child slot
		bvh4Node.ChildIndex[0] = primStart
		bvh4Node.PrimitiveCount[0] = int32(len(node.primitiveIndices))

		if node.box != nil {
			// Use conservative conversion to ensure we never miss intersections
			bvh4Node.MinX[0] = conservativeFloat32Min(node.box.Min().X)
			bvh4Node.MinY[0] = conservativeFloat32Min(node.box.Min().Y)
			bvh4Node.MinZ[0] = conservativeFloat32Min(node.box.Min().Z)
			bvh4Node.MaxX[0] = conservativeFloat32Max(node.box.Max().X)
			bvh4Node.MaxY[0] = conservativeFloat32Max(node.box.Max().Y)
			bvh4Node.MaxZ[0] = conservativeFloat32Max(node.box.Max().Z)
		}

		bvh.Nodes = append(bvh.Nodes, bvh4Node)
		return nodeIndex
	}

	// This is an internal node - collect up to 4 children by collapsing binary tree levels
	children := collectChildren(node, 4)

	// CRITICAL CHECK: If we got more than 4 children, we're losing geometry!
	if len(children) > 4 {
		log.Warnf("collectChildren returned %d children, but BVH4 can only handle 4! Geometry will be lost!", len(children))
	}

	// Reserve space for this node
	bvh.Nodes = append(bvh.Nodes, bvh4Node)

	// Process each child
	for i := 0; i < len(children) && i < 4; i++ {
		child := children[i]
		childIndex := flattenBVH4(child, bvh, primitiveIndices)

		bvh.Nodes[nodeIndex].ChildIndex[i] = childIndex

		if child.box != nil {
			// Use conservative conversion to ensure we never miss intersections
			bvh.Nodes[nodeIndex].MinX[i] = conservativeFloat32Min(child.box.Min().X)
			bvh.Nodes[nodeIndex].MinY[i] = conservativeFloat32Min(child.box.Min().Y)
			bvh.Nodes[nodeIndex].MinZ[i] = conservativeFloat32Min(child.box.Min().Z)
			bvh.Nodes[nodeIndex].MaxX[i] = conservativeFloat32Max(child.box.Max().X)
			bvh.Nodes[nodeIndex].MaxY[i] = conservativeFloat32Max(child.box.Max().Y)
			bvh.Nodes[nodeIndex].MaxZ[i] = conservativeFloat32Max(child.box.Max().Z)
		}
	}

	return nodeIndex
}

// collectChildren collects up to maxChildren children by flattening internal nodes
// It performs a breadth-first expansion of the tree, stopping when we have exactly maxChildren.
func collectChildren(node *buildNode, maxChildren int) []*buildNode {
	if node == nil || len(node.primitiveIndices) > 0 {
		return []*buildNode{node}
	}

	// Start with just the input node's children
	if len(node.children) == 0 {
		return []*buildNode{node}
	}

	result := make([]*buildNode, 0, maxChildren)

	// Initialize with direct children
	for _, child := range node.children {
		if child != nil {
			result = append(result, child)
		}
	}

	// Try to expand internal nodes until we have maxChildren
	expanded := true
	for expanded && len(result) < maxChildren {
		expanded = false

		// Find the first internal node we can expand
		for i := 0; i < len(result); i++ {
			current := result[i]

			// Skip if it's a leaf
			if len(current.primitiveIndices) > 0 {
				continue
			}

			// Skip if it has no children
			if len(current.children) == 0 {
				continue
			}

			// Calculate if we have room to expand this node
			numNewChildren := len(current.children)
			numAfterExpansion := len(result) - 1 + numNewChildren // -1 because we're replacing current

			if numAfterExpansion <= maxChildren {
				// We have room - expand this node
				// Remove current node and add its children
				result = append(result[:i], result[i+1:]...)
				result = append(result, current.children...)
				expanded = true
				break // Start over
			}
		}
	}

	// Ensure we don't return more than maxChildren
	if len(result) > maxChildren {
		result = result[:maxChildren]
	}

	return result
}
