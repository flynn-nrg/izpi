package hitable

import (
	"sort"
	"time"

	"github.com/flynn-nrg/izpi/internal/aabb"
	https://github.com/flynn-nrg/go-vfx/tree/main/math32
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"

	log "github.com/sirupsen/logrus"
)

// Ensure interface compliance.
var _ Hitable = (*BVHNode)(nil)

// BVHNode represents a bounding volume hierarchy node.
type BVHNode struct {
	left  Hitable
	right Hitable
	time0 float32
	time1 float32
	box   *aabb.AABB
}

func NewBVH(hitables []Hitable, time0 float32, time1 float32) *BVHNode {
	log.Infof("Building BVH with %v elements", len(hitables))
	startTime := time.Now()
	randomFunc := fastrandom.NewWithDefaults().float32
	bvh := newBVH(hitables, randomFunc, time0, time1)
	log.Infof("Completed BVH construction in %v", time.Since(startTime))
	return bvh
}

func newBVH(hitables []Hitable, randomFunc func() float32, time0 float32, time1 float32) *BVHNode {
	bn := &BVHNode{
		time0: time0,
		time1: time1,
	}

	// Select a random axis for splitting (0=x, 1=y, 2=z)
	axis := int(3 * randomFunc())

	switch axis {
	case 0:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			return aabb.BoxLessX(box0, box1)
		})

	case 1:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			return aabb.BoxLessY(box0, box1)
		})

	case 2:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			if box1, ok = hitables[j].BoundingBox(time0, time1); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			return aabb.BoxLessZ(box0, box1)
		})
	}

	switch len(hitables) {
	case 0:
		log.Error("Cannot create BVH node with no hitables")
		return nil
	case 1:
		// For a single object, use it as both left and right child
		// This is safe since we never modify the object
		bn.left = hitables[0]
		bn.right = hitables[0]
	case 2:
		// For two objects, put one on each side
		bn.left = hitables[0]
		bn.right = hitables[1]
	default:
		// For more than two objects, recursively split the list
		mid := len(hitables) / 2
		bn.left = newBVH(hitables[:mid], randomFunc, time0, time1)
		bn.right = newBVH(hitables[mid:], randomFunc, time0, time1)
	}

	// Compute the bounding box for this node
	var leftBox, rightBox *aabb.AABB
	var ok bool

	if leftBox, ok = bn.left.BoundingBox(time0, time1); !ok {
		log.Error("Left child has no bounding box in BVH node")
		return nil
	}
	if rightBox, ok = bn.right.BoundingBox(time0, time1); !ok {
		log.Error("Right child has no bounding box in BVH node")
		return nil
	}

	bn.box = aabb.SurroundingBox(leftBox, rightBox)
	return bn
}

func (bn *BVHNode) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	if bn.box.Hit(r, tMin, tMax) {
		leftRec, leftMat, hitLeft := bn.left.Hit(r, tMin, tMax)
		rightRec, rightMat, hitRight := bn.right.Hit(r, tMin, tMax)

		if hitLeft && hitRight {
			if leftRec.T() < rightRec.T() {
				return leftRec, leftMat, true
			}
			return rightRec, rightMat, true
		}

		if hitLeft {
			return leftRec, leftMat, true
		}

		if hitRight {
			return rightRec, rightMat, true
		}
	}

	return nil, nil, false
}

func (bn *BVHNode) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	if bn.box.Hit(r, tMin, tMax) {
		leftRec, hitLeft, hitLeftEdge := bn.left.HitEdge(r, tMin, tMax)
		rightRec, hitRight, hitRightEdge := bn.right.HitEdge(r, tMin, tMax)

		if hitLeft && hitRight {
			if leftRec.T() < rightRec.T() {
				return leftRec, true, hitLeftEdge
			}
			return rightRec, true, hitRightEdge
		}

		if hitLeft {
			return leftRec, true, hitLeftEdge
		}

		if hitRight {
			return rightRec, true, hitRightEdge
		}
	}

	return nil, false, false
}

func (bn *BVHNode) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return bn.box, true
}

func (bn *BVHNode) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32 {
	return 0.0
}

func (bn *BVHNode) Random(o *vec3.Vec3Impl, _ *fastrandom.LCG) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}

func (bn *BVHNode) IsEmitter() bool {
	return false
}
