package render

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWalkGridSpiral(t *testing.T) {
	testData := []struct {
		name  string
		sizeX int
		sizeY int
		want  []gridPos
	}{
		{
			name:  "3x3 grid",
			sizeX: 3,
			sizeY: 3,
			want: []gridPos{
				{X: 1, Y: 1},
				{X: 1},
				{X: 2},
				{X: 2, Y: 1},
				{X: 2, Y: 2},
				{X: 1, Y: 2},
				{Y: 2},
				{Y: 1},
				{},
			},
		},
		{
			name:  "4x4 grid",
			sizeX: 4,
			sizeY: 4,
			want: []gridPos{
				{X: 2, Y: 2},
				{X: 2, Y: 1},
				{X: 3, Y: 1},
				{X: 3, Y: 2},
				{X: 3, Y: 3},
				{X: 2, Y: 3},
				{X: 1, Y: 3},
				{X: 1, Y: 2},
				{X: 1, Y: 1},
				{X: 1},
				{X: 2},
				{X: 3},
				{},
				{Y: 1},
				{Y: 2},
				{Y: 3},
			},
		},
		{
			name:  "5x5 grid",
			sizeX: 5,
			sizeY: 5,
			want: []gridPos{
				{X: 2, Y: 2}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 3, Y: 2}, {X: 3, Y: 3},
				{X: 2, Y: 3}, {X: 1, Y: 3}, {X: 1, Y: 2}, {X: 1, Y: 1}, {X: 1}, {X: 2}, {X: 3},
				{X: 4}, {X: 4, Y: 1}, {X: 4, Y: 2}, {X: 4, Y: 3}, {X: 4, Y: 4}, {X: 3, Y: 4},
				{X: 2, Y: 4}, {X: 1, Y: 4}, {Y: 4}, {Y: 3}, {Y: 2}, {Y: 1}, {},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got := walkGridSpiral(test.sizeX, test.sizeY)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("WalkGridSpiral() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
