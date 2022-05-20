package mat3

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

func TestMatrixVectorMul(t *testing.T) {
	testData := []struct {
		name   string
		matrix *Mat3
		vector *vec3.Vec3Impl
		want   *vec3.Vec3Impl
	}{
		{
			name: "Multiply by identity",
			matrix: &Mat3{
				A11: 1,
				A22: 1,
				A33: 1,
			},
			vector: &vec3.Vec3Impl{X: 1, Y: 2, Z: 3},
			want:   &vec3.Vec3Impl{X: 1, Y: 2, Z: 3},
		},
		{
			name:   "TBN transformation for normal map vector Z = 1 for a triangle sitting on the XY plane",
			matrix: NewTBN(&vec3.Vec3Impl{X: -1}, &vec3.Vec3Impl{Y: 1}, &vec3.Vec3Impl{Z: -1}),
			vector: &vec3.Vec3Impl{Z: 1},
			want:   &vec3.Vec3Impl{Z: -1},
		},
		{
			name:   "TBN transformation for normal map vector Z = 1 for a triangle sitting on the XZ plane",
			matrix: NewTBN(&vec3.Vec3Impl{X: -1}, &vec3.Vec3Impl{Z: 1}, &vec3.Vec3Impl{Y: 1}),
			vector: &vec3.Vec3Impl{Z: 1},
			want:   &vec3.Vec3Impl{Y: 1},
		},
		{
			name:   "TBN transformation for normal map vector Z = 1 for a triangle sitting on the YZ plane",
			matrix: NewTBN(&vec3.Vec3Impl{Z: 1}, &vec3.Vec3Impl{Y: 1}, &vec3.Vec3Impl{X: -1}),
			vector: &vec3.Vec3Impl{Z: 1},
			want:   &vec3.Vec3Impl{X: -1},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := MatrixVectorMul(test.matrix, test.vector)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("MatrixVectorMul() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
