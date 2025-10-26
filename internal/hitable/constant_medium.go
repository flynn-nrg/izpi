package hitable

import (
	"github.com/flynn-nrg/go-vfx/math32"

	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/texture"
)

// Ensure interface compliance.
var _ Hitable = (*ConstantMedium)(nil)

// ConstantMedium represents a medium with constant density.
type ConstantMedium struct {
	hitable       Hitable
	density       float32
	phaseFunction material.Material
	random        *fastrandom.XorShift
}

// NewConstantMedium returns a new instance of the constant medium hitable.
func NewConstantMedium(hitable Hitable, density float32, a texture.Texture) *ConstantMedium {
	return &ConstantMedium{
		hitable:       hitable,
		density:       density,
		phaseFunction: material.NewIsotropic(a),
	}
}

func (cm *ConstantMedium) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	if rec1, _, ok := cm.hitable.Hit(r, -math32.MaxFloat32, math32.MaxFloat32); ok {
		if rec2, _, ok := cm.hitable.Hit(r, rec1.T()+0.0001, math32.MaxFloat32); ok {
			rec1t := rec1.T()
			rec2t := rec2.T()
			if rec1t < tMin {
				rec1t = tMin
			}
			if rec2t < tMax {
				rec2t = tMax
			}
			if rec1t >= rec2t {
				return nil, nil, false
			}
			if rec1t < 0 {
				rec1t = 0
			}

			distanceInsideBoundary := (rec2t - rec1t) * r.Direction().Length()
			hitDistance := -(1 / cm.density) * math32.Log(cm.random.Float32())
			if hitDistance < distanceInsideBoundary {
				t := rec1t + hitDistance/r.Direction().Length()
				// arbitrary
				normal := vec3.Vec3Impl{X: 1}
				hr := hitrecord.New(t, 0, 0, r.PointAtParameter(t), normal)
				return hr, cm.phaseFunction, true
			}
		}
	}

	return nil, nil, false
}

func (cm *ConstantMedium) HitEdge(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, bool, bool) {
	return nil, false, false
}

func (cm *ConstantMedium) BoundingBox(time0 float32, time1 float32) (*aabb.AABB, bool) {
	return cm.hitable.BoundingBox(time0, time1)
}

func (cm *ConstantMedium) PDFValue(o vec3.Vec3Impl, v vec3.Vec3Impl) float32 {
	return 0.0
}

func (cm *ConstantMedium) Random(o vec3.Vec3Impl, _ *fastrandom.XorShift) vec3.Vec3Impl {
	return vec3.Vec3Impl{X: 1}
}

func (cm *ConstantMedium) IsEmitter() bool {
	return cm.hitable.IsEmitter()
}
