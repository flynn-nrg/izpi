package segment

import (
	"testing"

	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/google/go-cmp/cmp"
)

func TestBelongs(t *testing.T) {
	testData := []struct {
		name string
		s    *Segment
		p    *vec3.Vec3Impl
		want bool
	}{
		{
			name: "Point on start of segment",
			s: &Segment{
				A: &vec3.Vec3Impl{X: 1, Y: 2, Z: 3},
				B: &vec3.Vec3Impl{X: 4, Y: 5, Z: 6},
			},
			p:    &vec3.Vec3Impl{X: 1, Y: 2, Z: 3},
			want: true,
		},
		{
			name: "Point on end of segment",
			s: &Segment{
				A: &vec3.Vec3Impl{X: 1, Y: 2, Z: 3},
				B: &vec3.Vec3Impl{X: 4, Y: 5, Z: 6},
			},
			p:    &vec3.Vec3Impl{X: 4, Y: 5, Z: 6},
			want: true,
		},
		{
			name: "Point is halfway",
			s: &Segment{
				A: &vec3.Vec3Impl{X: -1, Y: 0, Z: 0},
				B: &vec3.Vec3Impl{X: 1, Y: 1, Z: 4},
			},
			p:    &vec3.Vec3Impl{X: 0, Y: .5, Z: 2},
			want: true,
		},
		{
			name: "Point is not part of this segment",
			s: &Segment{
				A: &vec3.Vec3Impl{X: -1, Y: 0, Z: 0},
				B: &vec3.Vec3Impl{X: 1, Y: 1, Z: 4},
			},
			p: &vec3.Vec3Impl{X: 0, Y: .5, Z: 3},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := Belongs(test.s, test.p)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Belongs() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
