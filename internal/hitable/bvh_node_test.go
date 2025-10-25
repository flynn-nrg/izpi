package hitable

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/aabb"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/google/go-cmp/cmp"
)

func TestNewBVH(t *testing.T) {
	testData := []struct {
		name     string
		hitables []Hitable
		time0    float64
		time1    float64
		want     *BVHNode
	}{
		{
			name:     "A single sphere",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0)},
			time0:    0,
			time1:    1,
			want: &BVHNode{
				left: &Sphere{
					center0:  vec3.Vec3Impl{},
					center1:  vec3.Vec3Impl{},
					radius:   1,
					material: makeMaterial(),
				},
				right: &Sphere{
					center0:  vec3.Vec3Impl{},
					center1:  vec3.Vec3Impl{},
					radius:   1,
					material: makeMaterial(),
				},
				box: aabb.New(vec3.Vec3Impl{X: -1, Y: -1, Z: -1}, vec3.Vec3Impl{X: 1, Y: 1, Z: 1}),
			},
		},
		{
			name:     "Two spheres",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0), makeSphere(1, 0, 0, 1.0)},
			time0:    0,
			time1:    1,
			want: &BVHNode{
				left: &Sphere{
					center0:  vec3.Vec3Impl{},
					center1:  vec3.Vec3Impl{},
					radius:   1,
					material: makeMaterial(),
				},
				right: &Sphere{
					center0:  vec3.Vec3Impl{X: 1},
					center1:  vec3.Vec3Impl{X: 1},
					radius:   1,
					material: makeMaterial(),
				},
				box: aabb.New(vec3.Vec3Impl{X: -1, Y: -1, Z: -1}, vec3.Vec3Impl{X: 2, Y: 1, Z: 1}),
			},
		},
		{
			name:     "Five spheres",
			hitables: []Hitable{makeSphere(0, 0, 0, 1.0), makeSphere(1, 0, 0, 1.0), makeSphere(0, 1, 0, 1.0), makeSphere(1, 1, 0, 1.0), makeSphere(1, 1, 1, 1.0)},
			time0:    0,
			time1:    1,
			want: &BVHNode{
				left: &BVHNode{
					left: &Sphere{
						center0:  vec3.Vec3Impl{},
						center1:  vec3.Vec3Impl{},
						radius:   1,
						material: makeMaterial(),
					},
					right: &Sphere{
						center0:  vec3.Vec3Impl{Y: 1},
						center1:  vec3.Vec3Impl{Y: 1},
						radius:   1,
						material: makeMaterial(),
					},
					box: aabb.New(vec3.Vec3Impl{X: -1, Y: -1, Z: -1}, vec3.Vec3Impl{X: 1, Y: 2, Z: 1}),
				},
				right: &BVHNode{
					left: &BVHNode{
						left: &Sphere{
							center0:  vec3.Vec3Impl{X: 1},
							center1:  vec3.Vec3Impl{X: 1},
							radius:   1,
							material: makeMaterial(),
						},
						right: &Sphere{
							center0:  vec3.Vec3Impl{X: 1},
							center1:  vec3.Vec3Impl{X: 1},
							radius:   1,
							material: makeMaterial(),
						},
						box: aabb.New(vec3.Vec3Impl{X: 0, Y: -1, Z: -1}, vec3.Vec3Impl{X: 2, Y: 1, Z: 1}),
					},
					right: &BVHNode{
						left: &Sphere{
							center0:  vec3.Vec3Impl{X: 1, Y: 1},
							center1:  vec3.Vec3Impl{X: 1, Y: 1},
							radius:   1,
							material: makeMaterial(),
						},
						right: &Sphere{
							center0:  vec3.Vec3Impl{X: 1, Y: 1, Z: 1},
							center1:  vec3.Vec3Impl{X: 1, Y: 1, Z: 1},
							radius:   1,
							material: makeMaterial(),
						},
						box: aabb.New(vec3.Vec3Impl{X: 0, Y: 0, Z: -1}, vec3.Vec3Impl{X: 2, Y: 2, Z: 2}),
					},
					box: aabb.New(vec3.Vec3Impl{X: 0, Y: -1, Z: -1}, vec3.Vec3Impl{X: 2, Y: 2, Z: 2}),
				},

				box: aabb.New(vec3.Vec3Impl{X: -1, Y: -1, Z: -1}, vec3.Vec3Impl{X: 2, Y: 2, Z: 2}),
			},
		},
	}

	ramdonFunc := func() float64 { return 0 }
	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			got := newBVH(test.hitables, ramdonFunc, test.time0, test.time0)
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(BVHNode{}),
				cmp.AllowUnexported(Sphere{}),
				cmp.AllowUnexported(material.Lambertian{}),
				cmp.AllowUnexported(texture.Constant{}),
				cmp.AllowUnexported(aabb.AABB{})); diff != "" {
				t.Errorf("NewBVH() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func makeSphere(x float64, y float64, z float64, r float64) *Sphere {
	return NewSphere(
		vec3.Vec3Impl{
			X: x,
			Y: y,
			Z: z,
		},
		vec3.Vec3Impl{
			X: x,
			Y: y,
			Z: z,
		}, 0, 0, r, makeMaterial())
}

func makeMaterial() material.Material {
	return material.NewLambertian(texture.NewConstant(vec3.Vec3Impl{X: 0.2, Y: 0.3, Z: 0.1}))
}
