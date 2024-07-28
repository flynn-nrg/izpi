package hitable

import (
	//	"math/rand"
	"math/rand"
	"sort"
	"time"

	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/fastrandom"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"

	log "github.com/sirupsen/logrus"
)

// Ensure interface compliance.
var _ Hitable = (*BVHNode)(nil)

// BVHNode represents a bounding volume hierarchy node.
type BVHNode struct {
	left  Hitable
	right Hitable
	time0 float64
	time1 float64
	box   *aabb.AABB
}

func NewBVH(hitables []Hitable, time0 float64, time1 float64) *BVHNode {
	log.Infof("Building BVH with %v elements", len(hitables))
	startTime := time.Now()
	randomFunc := rand.Float64
	bvh := newBVH(hitables, randomFunc, time0, time1)
	log.Infof("Completed BVH construction in %v", time.Since(startTime))
	return bvh
}

func newBVH(hitables []Hitable, randomFunc func() float64, time0 float64, time1 float64) *BVHNode {
	bn := &BVHNode{
		time0: time0,
		time1: time1,
	}

	axis := int(3 * rand.Float64())

	switch axis {
	case 0:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(0, 0); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			if box1, ok = hitables[j].BoundingBox(0, 0); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			return aabb.BoxLessX(box0, box1)
		})

	case 1:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(0, 0); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			if box1, ok = hitables[j].BoundingBox(0, 0); !ok {
				log.Warning("no bounding box in BVH node")
				return false
			}
			return aabb.BoxLessY(box0, box1)
		})

	case 2:
		sort.Slice(hitables, func(i, j int) bool {
			var box0, box1 *aabb.AABB
			var ok bool
			if box0, ok = hitables[i].BoundingBox(0, 0); !ok {
				return false
			}
			if box1, ok = hitables[j].BoundingBox(0, 0); !ok {
				return false
			}
			return aabb.BoxLessZ(box0, box1)
		})

	}

	if len(hitables) == 1 {
		bn.left = hitables[0]
		bn.right = bn.left
	} else if len(hitables) == 2 {
		bn.left = hitables[0]
		bn.right = hitables[1]
	} else {
		bn.left = newBVH(hitables[:len(hitables)/2], randomFunc, time0, time1)
		bn.right = newBVH(hitables[len(hitables)/2:], randomFunc, time0, time1)
	}

	var leftBox, rightBox *aabb.AABB
	var ok bool
	if leftBox, ok = bn.left.BoundingBox(time0, time1); !ok {
		log.Warning("no bounding box in BVH node")
	}
	if rightBox, ok = bn.right.BoundingBox(time0, time1); !ok {
		log.Warning("no bounding box in BVH node")
	}

	bn.box = aabb.SurroundingBox(leftBox, rightBox)

	return bn
}

func (bn *BVHNode) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, material.Material, bool) {
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

func (bn *BVHNode) HitEdge(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, bool, bool) {
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

func (bn *BVHNode) BoundingBox(time0 float64, time1 float64) (*aabb.AABB, bool) {
	return bn.box, true
}

func (bn *BVHNode) PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64 {
	return 0.0
}

func (bn *BVHNode) Random(o *vec3.Vec3Impl, _ *fastrandom.LCG) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1}
}

func (bn *BVHNode) IsEmitter() bool {
	return false
}
