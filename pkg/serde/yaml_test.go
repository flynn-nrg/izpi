package serde

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSerialise(t *testing.T) {
	testData := []struct {
		name     string
		wantFile string
		wantErr  bool
		scene    *Scene
	}{
		{
			name:     "One camera, one sphere and one light",
			wantFile: "testdata/sphere.yaml",
			scene: &Scene{
				Name: "One sphere",
				Camera: Camera{
					LookFrom:  Vec3{X: 0, Y: 0, Z: -10},
					LookAt:    Vec3{X: 0, Y: 0, Z: 1},
					VUp:       Vec3{Y: 1},
					VFov:      60,
					Aperture:  1.8,
					FocusDist: 10,
					Time1:     1,
				},
				Objects: Objects{
					Spheres: []Sphere{
						{
							Center: Vec3{X: 0, Y: 5, Z: 0},
							Radius: 5,
							Material: Material{
								Name: "White",
								Type: LambertMaterial,
								Lambert: Lambert{
									Albedo: Texture{
										Name: "White",
										Type: "constant",
										Constant: Constant{
											Value: Vec3{X: .73, Y: .73, Z: .73},
										},
									},
								},
							},
						},
						{
							Center: Vec3{X: -50, Y: 50, Z: 3},
							Radius: 1,
							Material: Material{
								Name: "Green light",
								Type: DiffuseLightMaterial,
								DiffuseLight: DiffuseLight{
									Emit: Texture{
										Name: "Green",
										Type: "constant",
										Constant: Constant{
											Value: Vec3{Y: .7},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var gotBuff bytes.Buffer
			w := bufio.NewWriter(&gotBuff)
			r, err := os.Open(test.wantFile)
			if err != nil {
				t.Fatal(err)
			}
			wantBuf, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}

			y := &Yaml{}
			gotErr := y.Serialise(test.scene, w)
			w.Flush()
			if (gotErr != nil) != test.wantErr {
				t.Errorf("Test: %q :  Got error %v, wanted err=%v", test.name, gotErr, test.wantErr)
			}
			if diff := cmp.Diff(string(wantBuf), gotBuff.String()); diff != "" {
				t.Errorf("Serialise() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

func TestDeserialise(t *testing.T) {
	testData := []struct {
		name      string
		inputFile string
		want      *Scene
		wantErr   bool
	}{
		{
			name:      "One camera, one sphere and one light",
			inputFile: "testdata/sphere.yaml",
			want: &Scene{
				Name: "One sphere",
				Camera: Camera{
					LookFrom:  Vec3{X: 0, Y: 0, Z: -10},
					LookAt:    Vec3{X: 0, Y: 0, Z: 1},
					VUp:       Vec3{Y: 1},
					VFov:      60,
					Aperture:  1.8,
					FocusDist: 10,
					Time1:     1,
				},
				Objects: Objects{
					Spheres: []Sphere{
						{
							Center: Vec3{X: 0, Y: 5, Z: 0},
							Radius: 5,
							Material: Material{
								Name: "White",
								Type: LambertMaterial,
								Lambert: Lambert{
									Albedo: Texture{
										Name: "White",
										Type: "constant",
										Constant: Constant{
											Value: Vec3{X: .73, Y: .73, Z: .73},
										},
									},
								},
							},
						},
						{
							Center: Vec3{X: -50, Y: 50, Z: 3},
							Radius: 1,
							Material: Material{
								Name: "Green light",
								Type: DiffuseLightMaterial,
								DiffuseLight: DiffuseLight{
									Emit: Texture{
										Name: "Green",
										Type: "constant",
										Constant: Constant{
											Value: Vec3{Y: .7},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			r, err := os.Open(test.inputFile)
			if err != nil {
				t.Fatal(err)
			}
			y := &Yaml{}
			got, gotErr := y.Deserialise(r)
			if (gotErr != nil) != test.wantErr {
				t.Errorf("Test: %q :  Got error %v, wanted err=%v", test.name, gotErr, test.wantErr)
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Deserialise() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
