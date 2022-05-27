package displacement

import (
	"testing"

	"github.com/flynn-nrg/izpi/pkg/aabb"
	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
	"github.com/google/go-cmp/cmp"
)

func TestTessellate(t *testing.T) {
	testData := []struct {
		name  string
		input *minimalTriangle
		want  []*minimalTriangle
	}{
		{
			name: "Triangle on the XY plane",
			input: &minimalTriangle{
				vertex0: vec3.Vec3Impl{X: -1},
				vertex1: vec3.Vec3Impl{X: 1},
				vertex2: vec3.Vec3Impl{Y: 1},
				u1:      1.0,
				u2:      0.5,
				v2:      1.0,
			},
			want: []*minimalTriangle{

				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{},
					vertex2: vec3.Vec3Impl{X: -0.5, Y: 0.5},
					u1:      0.5,
					u2:      0.25,
					v2:      0.5,
				},
				{
					vertex0: vec3.Vec3Impl{},
					vertex1: vec3.Vec3Impl{X: 0.5, Y: 0.5},
					vertex2: vec3.Vec3Impl{X: -0.5, Y: 0.5},
					u0:      0.5,
					u1:      0.75,
					u2:      0.25,
					v1:      0.5,
					v2:      0.5,
				},
				{
					vertex0: vec3.Vec3Impl{},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{X: 0.5, Y: 0.5},
					u0:      0.5,
					u1:      1,
					u2:      0.75,
					v2:      0.5,
				},
				{
					vertex0: vec3.Vec3Impl{X: -0.5, Y: 0.5},
					vertex1: vec3.Vec3Impl{X: 0.5, Y: 0.5},
					vertex2: vec3.Vec3Impl{Y: 1},
					u0:      0.25,
					u1:      0.75,
					u2:      0.5,
					v0:      0.5,
					v1:      0.5,
					v2:      1,
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := tessellate(test.input)
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(minimalTriangle{})); diff != "" {
				t.Errorf("tessellate() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

func TestApplyTessellation(t *testing.T) {
	testData := []struct {
		name             string
		resU             int
		resV             int
		input            []*minimalTriangle
		wantNumTriangles int
	}{
		{
			name: "3x3 displacement map",
			resU: 3,
			resV: 3,
			input: []*minimalTriangle{
				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{Y: 1},
					u1:      1.0,
					u2:      0.5,
					v2:      1.0,
				},
			},
			wantNumTriangles: 4,
		},
		{
			name: "4x2 displacement map",
			resU: 4,
			resV: 2,
			input: []*minimalTriangle{
				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{Y: 1},
					u1:      1.0,
					u2:      0.5,
					v2:      1.0,
				},
			},
			wantNumTriangles: 16,
		},
		{
			name: "2x4 displacement map",
			resU: 2,
			resV: 4,
			input: []*minimalTriangle{
				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{Y: 1},
					u1:      1.0,
					u2:      0.5,
					v2:      1.0,
				},
			},
			wantNumTriangles: 16,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			maxDeltaU := 1.0 / float64(test.resU-1)
			maxDeltaV := 1.0 / float64(test.resV-1)
			got := applyTessellation(test.input, maxDeltaU, maxDeltaV)
			if len(got) != test.wantNumTriangles {
				t.Errorf("applyTessellation() mismatch: expected %v triangles, got %v\n", test.wantNumTriangles, len(got))
			}
		})
	}
}

func TestApplyDisplacement(t *testing.T) {
	testData := []struct {
		name    string
		texture texture.Texture
		min     float64
		max     float64
		input   []*minimalTriangle
		want    []*hitable.Triangle
	}{
		{
			name: "Triangle on the XZ plane",
			input: []*minimalTriangle{
				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{Z: 1},
					u1:      1.0,
					u2:      0.5,
					v2:      1.0,
				},
			},
			max:     1.0,
			texture: texture.NewConstant(&vec3.Vec3Impl{Z: 1}),
			want: []*hitable.Triangle{hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -1, Y: -1}, &vec3.Vec3Impl{X: 1, Y: -1}, &vec3.Vec3Impl{Y: -1, Z: 1},
				0, 0, 1, 0, 0.5, 1.0, nil)},
		},
		{
			name: "Triangle on the XZ plane. Different min and max settings.",
			input: []*minimalTriangle{
				{
					vertex0: vec3.Vec3Impl{X: -1},
					vertex1: vec3.Vec3Impl{X: 1},
					vertex2: vec3.Vec3Impl{Z: 1},
					u1:      1.0,
					u2:      0.5,
					v2:      1.0,
				},
			},
			min:     -0.5,
			max:     0.5,
			texture: texture.NewConstant(&vec3.Vec3Impl{X: 1, Y: 1, Z: 1}),
			want: []*hitable.Triangle{hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -1, Y: -.5}, &vec3.Vec3Impl{X: 1, Y: -.5}, &vec3.Vec3Impl{Y: -0.5, Z: 1},
				0, 0, 1, 0, 0.5, 1.0, nil)},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := applyDisplacement(test.input, test.texture, test.min, test.max)
			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(hitable.Triangle{}), cmp.AllowUnexported(aabb.AABB{})); diff != "" {
				t.Errorf("applyDisplacement() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
