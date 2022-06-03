// Package scene implements structures and methods to work with scenes.
package scene

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flynn-nrg/izpi/pkg/camera"
	"github.com/flynn-nrg/izpi/pkg/displacement"
	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/material"
	"github.com/flynn-nrg/izpi/pkg/serde"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
	"github.com/flynn-nrg/izpi/pkg/wavefront"
)

var (
	ErrInvalidTextureType  = errors.New("invalid texture type")
	ErrInvalidMaterialType = errors.New("invalid material type")
)

// Scene represents a scene with the world elements, lights and camera.
type Scene struct {
	World  *hitable.HitableSlice
	Lights *hitable.HitableSlice
	Camera *camera.Camera
}

// New returns a new scene instance.
func New(world *hitable.HitableSlice, lights *hitable.HitableSlice, camera *camera.Camera) *Scene {
	return &Scene{
		World:  world,
		Lights: lights,
		Camera: camera,
	}
}

// FromStruct returns the internal representation of a scene from YAML data.
func FromYAML(r io.Reader, aspectOverride float64) (*Scene, error) {
	y := &serde.Yaml{}
	sceneStruct, err := y.Deserialise(r)
	if err != nil {
		return nil, err
	}

	s, err := FromStruct(sceneStruct, aspectOverride)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// FromStruct returns the internal representation of a scene from struct data.
func FromStruct(sceneStruct *serde.Scene, aspectOverride float64) (*Scene, error) {
	hitables, err := objectsFromStruct(&sceneStruct.Objects)
	if err != nil {
		return nil, err
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	return &Scene{
		World:  hitable.NewSlice(hitables),
		Lights: hitable.NewSlice(lights),
		Camera: cameraFromStruct(&sceneStruct.Camera, aspectOverride),
	}, nil
}

func cameraFromStruct(cam *serde.Camera, aspectOverride float64) *camera.Camera {
	lookFrom := &vec3.Vec3Impl{
		X: cam.LookFrom.X,
		Y: cam.LookFrom.Y,
		Z: cam.LookFrom.Z,
	}

	lookAt := &vec3.Vec3Impl{
		X: cam.LookAt.X,
		Y: cam.LookAt.Y,
		Z: cam.LookAt.Z,
	}

	vup := &vec3.Vec3Impl{
		X: cam.VUp.X,
		Y: cam.VUp.Y,
		Z: cam.VUp.Z,
	}

	vfov := cam.VFov

	var aspect float64
	if aspectOverride != 0.0 {
		aspect = aspectOverride
	} else {
		aspect = cam.Aspect
	}

	aperture := cam.Aperture
	focusDist := cam.FocusDist
	time0 := cam.Time0
	time1 := cam.Time1

	return camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, focusDist, time0, time1)
}

func imageFromStruct(im *serde.Image) (texture.Texture, error) {
	f, err := os.Open(im.FileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if strings.HasSuffix(im.FileName, "png") {
		t, err := texture.NewFromPNG(f)
		if err != nil {
			return nil, err
		}
		return t, nil
	} else if strings.HasSuffix(im.FileName, "hdr") {
		t, err := texture.NewFromHDR(f)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, ErrInvalidTextureType
}

func textureFromStruct(tex *serde.Texture) (texture.Texture, error) {
	switch tex.Type {
	case serde.ConstantTexture:
		return texture.NewConstant(&vec3.Vec3Impl{
			X: tex.Constant.Value.X,
			Y: tex.Constant.Value.Y,
			Z: tex.Constant.Value.Z,
		}), nil
	case serde.ImageTexture:
		return imageFromStruct(&tex.Image)
	case serde.NoiseTexture:
		return texture.NewNoise(tex.Noise.Scale), nil
	}

	return nil, ErrInvalidTextureType

}

func materialFromStruct(mat *serde.Material) (material.Material, error) {
	switch mat.Type {
	case serde.LambertMaterial:
		albedo, err := textureFromStruct(&mat.Lambert.Albedo)
		if err != nil {
			return nil, err
		}
		return material.NewLambertian(albedo), nil
	case serde.DielectricMaterial:
		return material.NewDielectric(mat.Dielectric.RefIdx), nil
	case serde.DiffuseLightMaterial:
		emit, err := textureFromStruct(&mat.DiffuseLight.Emit)
		if err != nil {
			return nil, err
		}
		return material.NewDiffuseLight(emit), nil
	case serde.MetalMaterial:
		return material.NewMetal(&vec3.Vec3Impl{
			X: mat.Metal.Albedo.X,
			Y: mat.Metal.Albedo.Y,
			Z: mat.Metal.Albedo.Z,
		}, mat.Metal.Fuzz), nil
	case serde.IsotropicMaterial:
		albedo, err := textureFromStruct(&mat.Isotropic.Albedo)
		if err != nil {
			return nil, err
		}
		return material.NewIsotropic(albedo), nil
	}

	return nil, ErrInvalidMaterialType
}

func sphereFromStruct(sphere *serde.Sphere) ([]hitable.Hitable, error) {
	mat, err := materialFromStruct(&sphere.Material)
	if err != nil {
		return nil, err
	}

	center := &vec3.Vec3Impl{X: sphere.Center.X, Y: sphere.Center.Y, Z: sphere.Center.Z}
	return []hitable.Hitable{hitable.NewSphere(center, center, 0, 1, sphere.Radius, mat)}, nil
}

func triangleFromStruct(triangle *serde.Triangle) ([]hitable.Hitable, error) {
	mat, err := materialFromStruct(&triangle.Material)
	if err != nil {
		return nil, err
	}

	vertex0 := &vec3.Vec3Impl{X: triangle.Vertex0.X, Y: triangle.Vertex0.Y, Z: triangle.Vertex0.Z}
	vertex1 := &vec3.Vec3Impl{X: triangle.Vertex1.X, Y: triangle.Vertex1.Y, Z: triangle.Vertex1.Z}
	vertex2 := &vec3.Vec3Impl{X: triangle.Vertex2.X, Y: triangle.Vertex2.Y, Z: triangle.Vertex2.Z}
	u0 := triangle.U0
	v0 := triangle.V0
	u1 := triangle.U1
	v1 := triangle.V1
	u2 := triangle.U2
	v2 := triangle.V2

	t := hitable.NewTriangleWithUV(vertex0, vertex1, vertex2, u0, v0, u1, v1, u2, v2, mat)
	if triangle.Displacement.DisplacementMap.FileName == "" {
		return []hitable.Hitable{t}, nil
	}

	displacementMap, err := imageFromStruct(&triangle.Displacement.DisplacementMap)
	if err != nil {
		return nil, err
	}

	min := triangle.Displacement.Min
	max := triangle.Displacement.Max
	displaced, err := displacement.ApplyDisplacementMap([]*hitable.Triangle{t}, displacementMap, min, max)
	if err != nil {
		return nil, err
	}

	hitables := []hitable.Hitable{}
	for _, tri := range displaced {
		hitables = append(hitables, tri)
	}
	return hitables, nil
}

func meshFromStruct(mesh *serde.Mesh) ([]hitable.Hitable, error) {
	r, err := os.Open(mesh.WavefrontData)
	if err != nil {
		return nil, err
	}
	obj, err := wavefront.NewObjFromReader(r, filepath.Dir(mesh.WavefrontData),
		wavefront.IGNORE_MATERIALS, wavefront.IGNORE_NORMALS)
	if err != nil {
		return nil, err
	}

	translate := &vec3.Vec3Impl{X: mesh.Translate.X, Y: mesh.Translate.Y, Z: mesh.Translate.Z}
	scale := &vec3.Vec3Impl{X: mesh.Scale.X, Y: mesh.Scale.Y, Z: mesh.Scale.Z}

	if !vec3.Equals(scale, &vec3.Vec3Impl{}) {
		obj.Scale(scale)
	}

	if !vec3.Equals(translate, &vec3.Vec3Impl{}) {
		obj.Translate(translate)
	}

	mat, err := materialFromStruct(&mesh.Material)
	if err != nil {
		return nil, err
	}

	hitables := []hitable.Hitable{}
	for i := range obj.Groups {
		faces, err := obj.GroupToHitablesWithCustomMaterial(i, mat)
		if err != nil {
			return nil, err
		}
		bvh := hitable.NewBVH(faces, 0, 1)
		hitables = append(hitables, bvh)
	}

	return hitables, nil
}

func objectsFromStruct(objects *serde.Objects) ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}

	for _, mesh := range objects.Meshes {
		h, err := meshFromStruct(&mesh)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, h...)
	}

	for _, tri := range objects.Triangles {
		t, err := triangleFromStruct(&tri)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, t...)
	}

	for _, sphere := range objects.Spheres {
		s, err := sphereFromStruct(&sphere)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, s...)
	}

	return hitables, nil
}
