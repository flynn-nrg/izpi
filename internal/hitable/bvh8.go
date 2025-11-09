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
var _ Hitable = (*BVH8)(nil)

// BVH8Node represents a single node in the 8-way BVH
// All bounds are float32 for SIMD efficiency (potential AVX-512)
type BVH8Node struct {
	// Structure-of-Arrays layout: 8 children, 6 bounds each (8 * 6 * 4 = 192 bytes)
	MinX [8]float32
	MinY [8]float32
	MinZ [8]float32
	MaxX [8]float32
	MaxY [8]float32
	MaxZ [8]float32

	// ChildIndex: Index into Nodes array (inner node) or Primitives array (leaf)
	// -1 indicates invalid/unused child slot
	ChildIndex [8]int32

	// PrimitiveCount: > 0 for leaf (number of primitives), 0 for inner node
	PrimitiveCount [8]int32
}

// BVH8 represents an 8-way bounding volume hierarchy
type BVH8 struct {
	Nodes      []BVH8Node
	Primitives []Hitable
	time0      float64
	time1      float64
}

func (bvh *BVH8) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
	if len(bvh.Nodes) == 0 {
		return nil, nil, false
	}

	var bestHitRec *hitrecord.HitRecord
	var bestMat material.Material
	hitFound := false

	// Convert ray to float32 for SIMD
	rayInvDir := vec3.Vec3Impl{
		X: 1.0 / r.Direction().X,
		Y: 1.0 / r.Direction().Y,
		Z: 1.0 / r.Direction().Z,
	}
	rInvX, rInvY, rInvZ := float32(rayInvDir.X), float32(rayInvDir.Y), float32(rayInvDir.Z)
	rOrgX, rOrgY, rOrgZ := float32(r.Origin().X), float32(r.Origin().Y), float32(r.Origin().Z)

	// Stack-based traversal (8-way BVH has shallower depth than 4-way)
	const MAX_DEPTH = 64
	stack := [MAX_DEPTH]int32{}
	stackPtr := int32(0)
	currentNodeIndex := int32(0)

	for {
		if currentNodeIndex == -1 {
			break
		}

		if currentNodeIndex >= int32(len(bvh.Nodes)) {
			break
		}

		node := bvh.Nodes[currentNodeIndex]

		// Test ray against 8 AABBs simultaneously
		mask := RayAABB8_SIMD(
			&rOrgX, &rOrgY, &rOrgZ,
			&rInvX, &rInvY, &rInvZ,
			&node.MinX, &node.MinY, &node.MinZ,
			&node.MaxX, &node.MaxY, &node.MaxZ,
			float32(tMax),
		)

		nextChildIndex := int32(-1)

		// Process up to 8 children
		for i := int32(0); i < 8; i++ {
			if (mask>>i)&1 == 0 {
				continue
			}

			childIndex := node.ChildIndex[i]
			primitiveCount := node.PrimitiveCount[i]

			if childIndex == -1 {
				continue
			}

			if primitiveCount > 0 {
				// Leaf node: test primitives
				for p := int32(0); p < primitiveCount; p++ {
					primitive := bvh.Primitives[childIndex+p]
					if rec, mat, hit := primitive.Hit(r, tMin, tMax); hit {
						tMax = rec.T()
						bestHitRec = rec
						bestMat = mat
						hitFound = true
					}
				}
			} else {
				// Inner node: queue for traversal
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

	return bestHitRec, bestMat, hitFound
}

func (bvh *BVH8) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
	if len(bvh.Nodes) == 0 {
		return nil, false, false
	}

	var bestHitRec *hitrecord.HitRecord
	var bestHitEdge bool
	hitFound := false

	rayInvDir := vec3.Vec3Impl{
		X: 1.0 / r.Direction().X,
		Y: 1.0 / r.Direction().Y,
		Z: 1.0 / r.Direction().Z,
	}
	rInvX, rInvY, rInvZ := float32(rayInvDir.X), float32(rayInvDir.Y), float32(rayInvDir.Z)
	rOrgX, rOrgY, rOrgZ := float32(r.Origin().X), float32(r.Origin().Y), float32(r.Origin().Z)

	const MAX_DEPTH = 64
	stack := [MAX_DEPTH]int32{}
	stackPtr := int32(0)
	currentNodeIndex := int32(0)

	for {
		if currentNodeIndex == -1 {
			break
		}

		if currentNodeIndex >= int32(len(bvh.Nodes)) {
			break
		}

		node := bvh.Nodes[currentNodeIndex]

		mask := RayAABB8_SIMD(
			&rOrgX, &rOrgY, &rOrgZ,
			&rInvX, &rInvY, &rInvZ,
			&node.MinX, &node.MinY, &node.MinZ,
			&node.MaxX, &node.MaxY, &node.MaxZ,
			float32(tMax),
		)

		nextChildIndex := int32(-1)

		for i := int32(0); i < 8; i++ {
			if (mask>>i)&1 == 0 {
				continue
			}

			childIndex := node.ChildIndex[i]
			primitiveCount := node.PrimitiveCount[i]

			if childIndex == -1 {
				continue
			}

			if primitiveCount > 0 {
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

func (bvh *BVH8) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	if len(bvh.Nodes) == 0 {
		return nil, false
	}

	root := bvh.Nodes[0]

	// Find overall bounds across all 8 children
	minX := float64(root.MinX[0])
	minY := float64(root.MinY[0])
	minZ := float64(root.MinZ[0])
	maxX := float64(root.MaxX[0])
	maxY := float64(root.MaxY[0])
	maxZ := float64(root.MaxZ[0])

	for i := 1; i < 8; i++ {
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

func (bvh *BVH8) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float64 {
	return 0.0
}

func (bvh *BVH8) Random(o vec3.Vec3Impl, _ *fastrandom.LCG) vec3.Vec3Impl {
	return vec3.Vec3Impl{X: 1}
}

func (bvh *BVH8) IsEmitter() bool {
	return false
}

// DebugStats returns statistics about the BVH8 structure
func (bvh *BVH8) DebugStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["num_nodes"] = len(bvh.Nodes)
	stats["num_primitives"] = len(bvh.Primitives)

	leafNodes := 0
	innerNodes := 0
	totalLeafPrimitives := 0
	emptySlots := 0

	for _, node := range bvh.Nodes {
		hasLeaf := false
		hasInner := false
		for i := 0; i < 8; i++ {
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

	if len(bvh.Nodes) > 0 {
		root := bvh.Nodes[0]
		rootChildren := 0
		for i := 0; i < 8; i++ {
			if root.ChildIndex[i] != -1 {
				rootChildren++
			}
		}
		stats["root_children"] = rootChildren
	}

	return stats
}

// Validate checks the integrity of the BVH8 structure
func (bvh *BVH8) Validate() []string {
	var errors []string

	if len(bvh.Nodes) == 0 {
		errors = append(errors, "BVH8 has no nodes")
		return errors
	}

	if len(bvh.Primitives) == 0 {
		errors = append(errors, "BVH8 has no primitives")
		return errors
	}

	primitivesReferenced := 0
	visited := make(map[int32]bool)
	toVisit := []int32{0}

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

		for i := 0; i < 8; i++ {
			childIdx := node.ChildIndex[i]
			primCount := node.PrimitiveCount[i]

			if childIdx == -1 {
				continue
			}

			if primCount > 0 {
				for p := int32(0); p < primCount; p++ {
					idx := childIdx + p
					if idx < 0 || idx >= int32(len(bvh.Primitives)) {
						errors = append(errors, fmt.Sprintf("Invalid primitive index %d", idx))
					} else {
						primitivesReferenced++
					}
				}
			} else {
				toVisit = append(toVisit, childIdx)
			}
		}
	}

	if primitivesReferenced != len(bvh.Primitives) {
		errors = append(errors, fmt.Sprintf("Primitive count mismatch: %d referenced, %d in array",
			primitivesReferenced, len(bvh.Primitives)))
	}

	return errors
}

// RayAABB8_SIMD tests a ray against 8 AABBs simultaneously
// Returns an 8-bit mask where bit i is set if the ray hits AABB i
func RayAABB8_SIMD(
	rayOrgX, rayOrgY, rayOrgZ *float32,
	rayInvDirX, rayInvDirY, rayInvDirZ *float32,
	minX, minY, minZ *[8]float32,
	maxX, maxY, maxZ *[8]float32,
	tMax float32,
) uint8 {
	return rayAABB8_SIMD_impl(rayOrgX, rayOrgY, rayOrgZ,
		rayInvDirX, rayInvDirY, rayInvDirZ,
		minX, minY, minZ,
		maxX, maxY, maxZ,
		tMax)
}

// NewBVH8 builds an 8-way BVH from a list of hitables
func NewBVH8(hitables []Hitable, time0 float64, time1 float64) *BVH8 {
	log.Infof("Building BVH8 with %v elements", len(hitables))
	startTime := time.Now()
	randomFunc := fastrandom.NewWithDefaults().Float64
	bvh := newBVH8(hitables, randomFunc, time0, time1)
	log.Infof("Completed BVH8 construction in %v", time.Since(startTime))

	stats := bvh.DebugStats()
	log.Infof("BVH8 Stats: nodes=%v, primitives=%v, leaf_nodes=%v, inner_nodes=%v, total_leaf_prims=%v, empty_slots=%v, root_children=%v",
		stats["num_nodes"], stats["num_primitives"], stats["leaf_nodes"], stats["inner_nodes"],
		stats["total_leaf_primitives"], stats["empty_slots"], stats["root_children"])

	if errors := bvh.Validate(); len(errors) > 0 {
		log.Errorf("BVH8 validation failed with %d errors:", len(errors))
		for i, err := range errors {
			log.Errorf("  Error %d: %s", i+1, err)
			if i >= 9 {
				log.Errorf("  ... and %d more errors", len(errors)-10)
				break
			}
		}
	} else {
		log.Info("BVH8 validation passed")
	}

	return bvh
}

// buildNode8 is a helper struct for building BVH8
type buildNode8 struct {
	box              *aabb.AABB
	children         []*buildNode8
	primitiveIndices []int
}

func newBVH8(hitables []Hitable, randomFunc func() float64, time0 float64, time1 float64) *BVH8 {
	if len(hitables) == 0 {
		log.Error("Cannot create BVH8 with no hitables")
		return nil
	}

	bvh := &BVH8{
		time0:      time0,
		time1:      time1,
		Primitives: hitables,
	}

	indices := make([]int, len(hitables))
	for i := range indices {
		indices[i] = i
	}

	// Build binary BVH first, then collapse to 8-way
	root := buildBinaryBVHForBVH8(hitables, indices, randomFunc, time0, time1)

	bvh.Nodes = make([]BVH8Node, 0)
	primitiveIndices := make([]int, 0)

	flattenBVH8(root, bvh, &primitiveIndices)

	// Reorder primitives
	reorderedPrimitives := make([]Hitable, len(primitiveIndices))
	for i, idx := range primitiveIndices {
		reorderedPrimitives[i] = hitables[idx]
	}
	bvh.Primitives = reorderedPrimitives

	return bvh
}

func buildBinaryBVHForBVH8(hitables []Hitable, indices []int, randomFunc func() float64, time0 float64, time1 float64) *buildNode8 {
	if len(hitables) == 0 {
		return nil
	}

	node := &buildNode8{}

	if len(hitables) == 1 {
		node.primitiveIndices = indices
		if box, ok := hitables[0].BoundingBox(time0, time1); ok {
			node.box = box
		}
		return node
	}

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

	axis := int(3 * randomFunc())

	workingHitables := make([]Hitable, len(hitables))
	workingIndices := make([]int, len(indices))
	copy(workingHitables, hitables)
	copy(workingIndices, indices)

	sortHitablesWithIndicesForBVH8(workingHitables, workingIndices, axis, time0, time1)

	if len(workingHitables) <= 8 {
		node.primitiveIndices = workingIndices
		return node
	}

	mid := len(workingHitables) / 2

	leftChild := buildBinaryBVHForBVH8(workingHitables[:mid], workingIndices[:mid], randomFunc, time0, time1)
	rightChild := buildBinaryBVHForBVH8(workingHitables[mid:], workingIndices[mid:], randomFunc, time0, time1)

	node.children = []*buildNode8{leftChild, rightChild}
	return node
}

func sortHitablesWithIndicesForBVH8(hitables []Hitable, indices []int, axis int, time0 float64, time1 float64) {
	type pair struct {
		hitable Hitable
		index   int
	}

	pairs := make([]pair, len(hitables))
	for i := range hitables {
		pairs[i] = pair{hitables[i], indices[i]}
	}

	switch axis {
	case 0:
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
	case 1:
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
	case 2:
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

	for i := range pairs {
		hitables[i] = pairs[i].hitable
		indices[i] = pairs[i].index
	}
}

func flattenBVH8(node *buildNode8, bvh *BVH8, primitiveIndices *[]int) int32 {
	if node == nil {
		return -1
	}

	nodeIndex := int32(len(bvh.Nodes))
	bvh8Node := BVH8Node{}

	// Initialize all children as invalid
	for i := 0; i < 8; i++ {
		bvh8Node.ChildIndex[i] = -1
		bvh8Node.PrimitiveCount[i] = 0
		bvh8Node.MinX[i] = math.MaxFloat32
		bvh8Node.MinY[i] = math.MaxFloat32
		bvh8Node.MinZ[i] = math.MaxFloat32
		bvh8Node.MaxX[i] = math.MaxFloat32
		bvh8Node.MaxY[i] = math.MaxFloat32
		bvh8Node.MaxZ[i] = math.MaxFloat32
	}

	if len(node.primitiveIndices) > 0 {
		primStart := int32(len(*primitiveIndices))
		*primitiveIndices = append(*primitiveIndices, node.primitiveIndices...)

		bvh8Node.ChildIndex[0] = primStart
		bvh8Node.PrimitiveCount[0] = int32(len(node.primitiveIndices))

		if node.box != nil {
			bvh8Node.MinX[0] = conservativeFloat32Min(node.box.Min().X)
			bvh8Node.MinY[0] = conservativeFloat32Min(node.box.Min().Y)
			bvh8Node.MinZ[0] = conservativeFloat32Min(node.box.Min().Z)
			bvh8Node.MaxX[0] = conservativeFloat32Max(node.box.Max().X)
			bvh8Node.MaxY[0] = conservativeFloat32Max(node.box.Max().Y)
			bvh8Node.MaxZ[0] = conservativeFloat32Max(node.box.Max().Z)
		}

		bvh.Nodes = append(bvh.Nodes, bvh8Node)
		return nodeIndex
	}

	// Collect up to 8 children by collapsing binary tree levels
	children := collectChildrenForBVH8(node, 8)

	if len(children) > 8 {
		log.Warnf("collectChildrenForBVH8 returned %d children, truncating to 8", len(children))
	}

	bvh.Nodes = append(bvh.Nodes, bvh8Node)

	for i := 0; i < len(children) && i < 8; i++ {
		child := children[i]
		childIndex := flattenBVH8(child, bvh, primitiveIndices)

		bvh.Nodes[nodeIndex].ChildIndex[i] = childIndex

		if child.box != nil {
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

func collectChildrenForBVH8(node *buildNode8, maxChildren int) []*buildNode8 {
	if node == nil || len(node.primitiveIndices) > 0 {
		return []*buildNode8{node}
	}

	if len(node.children) == 0 {
		return []*buildNode8{node}
	}

	result := make([]*buildNode8, 0, maxChildren)

	for _, child := range node.children {
		if child != nil {
			result = append(result, child)
		}
	}

	expanded := true
	for expanded && len(result) < maxChildren {
		expanded = false

		for i := 0; i < len(result); i++ {
			current := result[i]

			if len(current.primitiveIndices) > 0 {
				continue
			}

			if len(current.children) == 0 {
				continue
			}

			numNewChildren := len(current.children)
			numAfterExpansion := len(result) - 1 + numNewChildren

			if numAfterExpansion <= maxChildren {
				result = append(result[:i], result[i+1:]...)
				result = append(result, current.children...)
				expanded = true
				break
			}
		}
	}

	if len(result) > maxChildren {
		result = result[:maxChildren]
	}

	return result
}

