package hitable

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

func TestTriangleHit(t *testing.T) {
	testData := []struct {
		name    string
		vertex0 *vec3.Vec3Impl
		vertex1 *vec3.Vec3Impl
		vertex2 *vec3.Vec3Impl
		r       ray.Ray
		tMin    float64
		tMax    float64
		wantHR  *hitrecord.HitRecord
		wantHit bool
	}{
		{
			name:    "Triangle is parallel to ray",
			vertex0: &vec3.Vec3Impl{X: 1, Y: 0, Z: -1},
			vertex1: &vec3.Vec3Impl{X: 1, Y: 1, Z: -1},
			vertex2: &vec3.Vec3Impl{X: 0, Y: 0, Z: -1},
			r:       ray.New(&vec3.Vec3Impl{Y: -1}, &vec3.Vec3Impl{Y: 1}, 0),
			tMax:    math.MaxFloat64,
		},
		{
			name:    "Ray is perpendicular and hits triangle",
			vertex0: &vec3.Vec3Impl{X: .5, Y: -.5, Z: -10},
			vertex1: &vec3.Vec3Impl{X: -.5, Y: .5, Z: -10},
			vertex2: &vec3.Vec3Impl{X: 0, Y: .5, Z: -10},
			r:       ray.New(&vec3.Vec3Impl{Z: 1}, &vec3.Vec3Impl{Z: -1}, 0),
			tMax:    math.MaxFloat64,
			wantHR:  hitrecord.New(11, 0, 0, &vec3.Vec3Impl{Z: -10}, &vec3.Vec3Impl{Z: -1}),
			wantHit: true,
		},
		{
			name:    "Ray is perpendicular but does not hit triangle",
			vertex0: &vec3.Vec3Impl{X: .5, Y: -.5, Z: -10},
			vertex1: &vec3.Vec3Impl{X: -.5, Y: .5, Z: -10},
			vertex2: &vec3.Vec3Impl{X: 0, Y: .5, Z: -10},
			r:       ray.New(&vec3.Vec3Impl{X: -1, Y: 0, Z: 1}, &vec3.Vec3Impl{X: -1, Y: 0, Z: -1}, 0),
			tMax:    math.MaxFloat64,
		},
		{
			name:    "Ray hits triangle at an angle",
			vertex0: &vec3.Vec3Impl{X: .5, Y: -.5, Z: -20},
			vertex1: &vec3.Vec3Impl{X: -.5, Y: .5, Z: -10},
			vertex2: &vec3.Vec3Impl{X: -0.5, Y: -.5, Z: -10},
			r:       ray.New(&vec3.Vec3Impl{Z: 1}, &vec3.Vec3Impl{Z: -1}, 0),
			tMax:    math.MaxFloat64,
			wantHR:  hitrecord.New(16, 0, 0, &vec3.Vec3Impl{Z: -15}, &vec3.Vec3Impl{X: 0.9950371902099892, Z: 0.09950371902099892}),
			wantHit: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			tri := NewTriangle(test.vertex0, test.vertex1, test.vertex2, nil)
			gotHR, _, gotHit := tri.Hit(test.r, test.tMin, test.tMax)
			if diff := cmp.Diff(test.wantHit, gotHit); diff != "" {
				t.Errorf("Hit() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(test.wantHR, gotHR, cmp.AllowUnexported(hitrecord.HitRecord{})); diff != "" {
				t.Errorf("Hit() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
