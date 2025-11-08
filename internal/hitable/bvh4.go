package hitable

import (
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

// RayAABB4_SIMD tests a ray against 4 AABBs simultaneously.
// Returns a 4-bit mask where bit i is set if the ray hits AABB i.
// This is the pure Go version; assembly versions for ARM64 and AMD64 will be added later.
func RayAABB4_SIMD(
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
	return bvh
}

// buildNode is a helper struct for building the BVH4 tree
type buildNode struct {
	box        *aabb.AABB
	children   []*buildNode
	primitives []Hitable
	primStart  int32
	primCount  int32
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

	// Build a traditional binary BVH first
	root := buildBinaryBVH(hitables, randomFunc, time0, time1)

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
func buildBinaryBVH(hitables []Hitable, randomFunc func() float64, time0 float64, time1 float64) *buildNode {
	if len(hitables) == 0 {
		return nil
	}

	node := &buildNode{}

	// Compute bounding box for all primitives
	if len(hitables) == 1 {
		node.primitives = hitables
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

	// Sort along the chosen axis
	sortHitables(hitables, axis, time0, time1)

	if len(hitables) <= 4 {
		// Create leaf node
		node.primitives = hitables
		return node
	}

	// Split in the middle
	mid := len(hitables) / 2
	leftChild := buildBinaryBVH(hitables[:mid], randomFunc, time0, time1)
	rightChild := buildBinaryBVH(hitables[mid:], randomFunc, time0, time1)

	node.children = []*buildNode{leftChild, rightChild}
	return node
}

// sortHitables sorts hitables along the specified axis
func sortHitables(hitables []Hitable, axis int, time0 float64, time1 float64) {
	switch axis {
	case 0: // X
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessX(box0, box1)
		})
	case 1: // Y
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessY(box0, box1)
		})
	case 2: // Z
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				return false
			}
			return aabb.BoxLessZ(box0, box1)
		})
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
	for i := 0; i < 4; i++ {
		bvh4Node.ChildIndex[i] = -1
		bvh4Node.PrimitiveCount[i] = 0
		bvh4Node.MinX[i] = math.MaxFloat32
		bvh4Node.MinY[i] = math.MaxFloat32
		bvh4Node.MinZ[i] = math.MaxFloat32
		bvh4Node.MaxX[i] = -math.MaxFloat32
		bvh4Node.MaxY[i] = -math.MaxFloat32
		bvh4Node.MaxZ[i] = -math.MaxFloat32
	}

	// Check if this is a leaf node
	if len(node.primitives) > 0 {
		// This is a leaf - store primitives
		primStart := int32(len(*primitiveIndices))

		// Add primitive indices
		for _, prim := range node.primitives {
			// Find the index of this primitive in the original list
			for idx, p := range bvh.Primitives {
				if p == prim {
					*primitiveIndices = append(*primitiveIndices, idx)
					break
				}
			}
		}

		// Store in first child slot
		bvh4Node.ChildIndex[0] = primStart
		bvh4Node.PrimitiveCount[0] = int32(len(node.primitives))

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
func collectChildren(node *buildNode, maxChildren int) []*buildNode {
	if node == nil || len(node.primitives) > 0 {
		return []*buildNode{node}
	}

	result := make([]*buildNode, 0, maxChildren)
	queue := []*buildNode{node}

	for len(queue) > 0 && len(result) < maxChildren {
		current := queue[0]
		queue = queue[1:]

		if current == nil {
			continue
		}

		// If it's a leaf, add it to results
		if len(current.primitives) > 0 {
			result = append(result, current)
			continue
		}

		// If it's an internal node and we have space, expand it
		if len(current.children) > 0 {
			// Check if expanding would exceed limit
			if len(result)+len(current.children) <= maxChildren {
				queue = append(queue, current.children...)
			} else {
				// Can't expand further, treat as a subtree
				result = append(result, current)
			}
		}
	}

	return result
}
