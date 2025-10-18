package wavefront

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/google/go-cmp/cmp"
)

func TestNewObjFromReader(t *testing.T) {
	testData := []struct {
		name     string
		dataFile string
		want     *WavefrontObj
		wantErr  bool
	}{
		{
			name:     "A simple cube",
			dataFile: "testdata/cube.obj",
			want: &WavefrontObj{
				ObjectName: "Cube",
				HasNormals: true,
				HasUV:      true,
				Centre:     &vec3.Vec3Impl{},
				Vertices: []*vec3.Vec3Impl{
					{X: -0.5, Y: -0.5, Z: -0.5}, {X: 0.5, Y: -0.5, Z: -0.5},
					{X: 0.5, Y: -0.5, Z: 0.5}, {X: -0.5, Y: -0.5, Z: 0.5},
					{X: -0.5, Y: 0.5, Z: -0.5}, {X: 0.5, Y: 0.5, Z: -0.5},
					{X: 0.5, Y: 0.5, Z: 0.5}, {X: -0.5, Y: 0.5, Z: 0.5},
				},
				VertexNormals: []*vec3.Vec3Impl{
					{Y: -1}, {Z: -1}, {X: 1},
					{Z: 1}, {X: -1}, {Y: 1}},
				VertexUV: []*texture.UV{
					{U: 0.25}, {U: 0.5},
					{U: 0.25, V: 0.333333}, {U: 0.5, V: 0.333333},
					{U: 1, V: 0.666667}, {U: 0.75, V: 0.666667},
					{U: 1, V: 0.333333}, {U: 0.75, V: 0.333333},
					{U: 0.5, V: 0.666667}, {U: 0.25, V: 0.666667},
					{V: 0.666667}, {V: 0.333333},
					{U: 0.25, V: 1}, {U: 0.5, V: 1}},
				MtlLib: map[string]*Material{
					"Material1": {
						Name:  "Material1",
						Kd:    []float32{0.48, 0.48, 0.48},
						Ns:    256,
						D:     1,
						Illum: 2,
						Ka:    []float32{0, 0, 0},
						Ks:    []float32{0.04, 0.04, 0.04},
					},
				},

				Groups: []*Group{
					{
						Name:     "Cube1",
						FaceType: OBJ_FACE_TYPE_POLYGON,
						Material: "Material1",
						Faces: []*Face{
							{[]*VertexIndices{{VIdx: 1, VtIdx: 1, VnIdx: 1}, {VIdx: 2, VtIdx: 2, VnIdx: 1}, {VIdx: 4, VtIdx: 3, VnIdx: 1}}},
							{[]*VertexIndices{{VIdx: 2, VtIdx: 2, VnIdx: 1}, {VIdx: 3, VtIdx: 4, VnIdx: 1}, {VIdx: 4, VtIdx: 3, VnIdx: 1}}},
							{[]*VertexIndices{{VIdx: 5, VtIdx: 5, VnIdx: 2}, {VIdx: 6, VtIdx: 6, VnIdx: 2}, {VIdx: 1, VtIdx: 7, VnIdx: 2}}},
							{[]*VertexIndices{{VIdx: 6, VtIdx: 6, VnIdx: 2}, {VIdx: 2, VtIdx: 8, VnIdx: 2}, {VIdx: 1, VtIdx: 7, VnIdx: 2}}},
							{[]*VertexIndices{{VIdx: 6, VtIdx: 6, VnIdx: 3}, {VIdx: 7, VtIdx: 9, VnIdx: 3}, {VIdx: 2, VtIdx: 8, VnIdx: 3}}},
							{[]*VertexIndices{{VIdx: 7, VtIdx: 9, VnIdx: 3}, {VIdx: 3, VtIdx: 4, VnIdx: 3}, {VIdx: 2, VtIdx: 8, VnIdx: 3}}},
							{[]*VertexIndices{{VIdx: 7, VtIdx: 9, VnIdx: 4}, {VIdx: 8, VtIdx: 10, VnIdx: 4}, {VIdx: 3, VtIdx: 4, VnIdx: 4}}},
							{[]*VertexIndices{{VIdx: 8, VtIdx: 10, VnIdx: 4}, {VIdx: 4, VtIdx: 3, VnIdx: 4}, {VIdx: 3, VtIdx: 4, VnIdx: 4}}},
							{[]*VertexIndices{{VIdx: 8, VtIdx: 10, VnIdx: 5}, {VIdx: 5, VtIdx: 11, VnIdx: 5}, {VIdx: 4, VtIdx: 3, VnIdx: 5}}},
							{[]*VertexIndices{{VIdx: 5, VtIdx: 11, VnIdx: 5}, {VIdx: 1, VtIdx: 12, VnIdx: 5}, {VIdx: 4, VtIdx: 3, VnIdx: 5}}},
							{[]*VertexIndices{{VIdx: 5, VtIdx: 13, VnIdx: 6}, {VIdx: 8, VtIdx: 10, VnIdx: 6}, {VIdx: 6, VtIdx: 14, VnIdx: 6}}},
							{[]*VertexIndices{{VIdx: 8, VtIdx: 10, VnIdx: 6}, {VIdx: 7, VtIdx: 9, VnIdx: 6}, {VIdx: 6, VtIdx: 14, VnIdx: 6}}},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			file, err := os.Open(test.dataFile)
			if err != nil {
				t.Fatal(err)
			}

			got, gotErr := NewObjFromReader(file, filepath.Dir(test.dataFile))
			if (gotErr != nil) != test.wantErr {
				t.Errorf("Test: %q :  Got error %v, wanted err=%v", test.name, gotErr, test.wantErr)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("NewObjFromReader() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
