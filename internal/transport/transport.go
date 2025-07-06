package transport

import (
	"fmt"

	"github.com/flynn-nrg/izpi/internal/camera"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/material"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

type Transport struct {
	aspectOverride float64
	protoScene     *pb_transport.Scene
	triangles      []*pb_transport.Triangle
	textures       map[string]*texture.ImageTxt
	materials      map[string]material.Material
}

func NewTransport(aspectOverride float64, protoScene *pb_transport.Scene, triangles []*pb_transport.Triangle, textures map[string]*texture.ImageTxt) *Transport {
	return &Transport{aspectOverride: aspectOverride, protoScene: protoScene, triangles: triangles, textures: textures}
}

func (t *Transport) ToScene() (*scene.Scene, error) {
	camera := t.toSceneCamera(t.aspectOverride)

	materials, err := t.toSceneMaterials()
	if err != nil {
		return nil, err
	}
	t.materials = materials

	fmt.Println(t.materials)

	hitables, err := t.toSceneObjects()
	if err != nil {
		return nil, err
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	return &scene.Scene{
		World:  hitable.NewSlice([]hitable.Hitable{hitable.NewBVH(hitables, 0, 1)}),
		Lights: hitable.NewSlice(lights),
		Camera: camera,
	}, nil
}

func (t *Transport) toSceneMaterials() (map[string]material.Material, error) {
	materials := make(map[string]material.Material)

	for _, material := range t.protoScene.GetMaterials() {
		switch material.GetType() {
		case pb_transport.MaterialType_LAMBERT:
			lambert, err := t.toSceneLambertMaterial(material)
			if err != nil {
				return nil, err
			}
			materials[material.GetName()] = lambert
		case pb_transport.MaterialType_DIELECTRIC:
			dielectric, err := t.toSceneDielectricMaterial(material)
			if err != nil {
				return nil, err
			}
			materials[material.GetName()] = dielectric
		case pb_transport.MaterialType_DIFFUSE_LIGHT:
			diffuselight, err := t.toSceneDiffuseLightMaterial(material)
			if err != nil {
				return nil, err
			}
			materials[material.GetName()] = diffuselight
		case pb_transport.MaterialType_METAL:
			metal, err := t.toSceneMetalMaterial(material)
			if err != nil {
				return nil, err
			}
			materials[material.GetName()] = metal
		case pb_transport.MaterialType_PBR:
			pbr, err := t.toScenePBRMaterial(material)
			if err != nil {
				return nil, err
			}
			materials[material.GetName()] = pbr
		}
	}

	return materials, nil
}

func (t *Transport) toScenePBRMaterial(mat *pb_transport.Material) (material.Material, error) {
	pbr := mat.GetPbr()

	albedo, err := t.toSceneTexture(pbr.GetAlbedo())
	if err != nil {
		return nil, err
	}

	roughness, err := t.toSceneTexture(pbr.GetRoughness())
	if err != nil {
		return nil, err
	}

	metalness, err := t.toSceneTexture(pbr.GetMetalness())
	if err != nil {
		return nil, err
	}

	normalMap, err := t.toSceneTexture(pbr.GetNormalMap())
	if err != nil {
		return nil, err
	}

	sss, err := t.toSceneTexture(pbr.GetSss())
	if err != nil {
		return nil, err
	}

	sssRadius := float64(pbr.GetSssRadius())

	return material.NewPBR(albedo, roughness, metalness, normalMap, sss, sssRadius), nil
}

func (t *Transport) toSceneMetalMaterial(mat *pb_transport.Material) (material.Material, error) {
	metal := mat.GetMetal()

	albedo := &vec3.Vec3Impl{
		X: float64(metal.GetAlbedo().GetX()),
		Y: float64(metal.GetAlbedo().GetY()),
		Z: float64(metal.GetAlbedo().GetZ()),
	}

	fuzz := float64(metal.GetFuzz())

	return material.NewMetal(albedo, fuzz), nil
}

func (t *Transport) toSceneLambertMaterial(mat *pb_transport.Material) (material.Material, error) {
	lambert := mat.GetLambert()

	albedo, err := t.toSceneTexture(lambert.GetAlbedo())
	if err != nil {
		return nil, err
	}

	return material.NewLambertian(albedo), nil
}

func (t *Transport) toSceneDielectricMaterial(mat *pb_transport.Material) (material.Material, error) {
	dielectric := mat.GetDielectric()

	refidx := float64(dielectric.GetRefidx())

	return material.NewDielectric(refidx), nil
}

func (t *Transport) toSceneDiffuseLightMaterial(mat *pb_transport.Material) (material.Material, error) {
	diffuselight := mat.GetDiffuselight()

	emit, err := t.toSceneTexture(diffuselight.GetEmit())
	if err != nil {
		return nil, err
	}

	return material.NewDiffuseLight(emit), nil
}

func (t *Transport) toSceneTexture(text *pb_transport.Texture) (texture.Texture, error) {
	switch text.GetTextureProperties().(type) {
	case *pb_transport.Texture_Constant:
		return t.toSceneConstantTexture(text)
	case *pb_transport.Texture_Image:
		return t.toSceneImageTexture(text)
	}

	return nil, fmt.Errorf("unknown texture type: %T", text.GetTextureProperties())
}

func (t *Transport) toSceneConstantTexture(text *pb_transport.Texture) (texture.Texture, error) {
	constant := text.GetConstant()

	return texture.NewConstant(&vec3.Vec3Impl{
		X: float64(constant.GetValue().GetX()),
		Y: float64(constant.GetValue().GetY()),
		Z: float64(constant.GetValue().GetZ()),
	}), nil
}

func (t *Transport) toSceneImageTexture(text *pb_transport.Texture) (texture.Texture, error) {
	image := text.GetImage()

	imageText, ok := t.textures[image.GetFilename()]
	if !ok {
		return nil, fmt.Errorf("texture %s not found", image.GetFilename())
	}

	return imageText, nil
}

func (t *Transport) toSceneCamera(aspectOverride float64) *camera.Camera {
	protoCamera := t.protoScene.GetCamera()

	lookFrom := &vec3.Vec3Impl{
		X: float64(protoCamera.GetLookfrom().GetX()),
		Y: float64(protoCamera.GetLookfrom().GetY()),
		Z: float64(protoCamera.GetLookfrom().GetZ()),
	}

	lookAt := &vec3.Vec3Impl{
		X: float64(protoCamera.GetLookat().GetX()),
		Y: float64(protoCamera.GetLookat().GetY()),
		Z: float64(protoCamera.GetLookat().GetZ()),
	}

	vup := &vec3.Vec3Impl{
		X: float64(protoCamera.GetVup().GetX()),
		Y: float64(protoCamera.GetVup().GetY()),
		Z: float64(protoCamera.GetVup().GetZ()),
	}

	vfov := float64(protoCamera.GetVfov())

	var aspect float64
	if aspectOverride != 0.0 {
		aspect = aspectOverride
	} else {
		aspect = float64(protoCamera.GetAspect())
	}

	aperture := float64(protoCamera.GetAperture())
	focusDist := float64(protoCamera.GetFocusdist())
	time0 := float64(protoCamera.GetTime0())
	time1 := float64(protoCamera.GetTime1())

	return camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, focusDist, time0, time1)
}

func (t *Transport) toSceneObjects() ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}

	triangles, err := t.toSceneTriangles()
	if err != nil {
		return nil, err
	}
	hitables = append(hitables, triangles...)

	spheres, err := t.toSceneSpheres()
	if err != nil {
		return nil, err
	}
	hitables = append(hitables, spheres...)

	return hitables, nil
}

func (t *Transport) toSceneTriangles() ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}

	// Embedded triangles
	for _, triangle := range t.protoScene.GetObjects().GetTriangles() {
		tri, err := t.toSceneTriangle(triangle)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, tri)
	}

	// Streamed triangles
	for _, triangle := range t.triangles {
		tri, err := t.toSceneTriangle(triangle)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, tri)
	}

	return hitables, nil
}

func (t *Transport) toSceneTriangle(triangle *pb_transport.Triangle) (*hitable.Triangle, error) {
	material, ok := t.materials[triangle.GetMaterialName()]
	if !ok {
		return nil, fmt.Errorf("material %s not found", triangle.GetMaterialName())
	}

	vertex0 := &vec3.Vec3Impl{
		X: float64(triangle.GetVertex0().GetX()),
		Y: float64(triangle.GetVertex0().GetY()),
		Z: float64(triangle.GetVertex0().GetZ()),
	}

	vertex1 := &vec3.Vec3Impl{
		X: float64(triangle.GetVertex1().GetX()),
		Y: float64(triangle.GetVertex1().GetY()),
		Z: float64(triangle.GetVertex1().GetZ()),
	}

	vertex2 := &vec3.Vec3Impl{
		X: float64(triangle.GetVertex2().GetX()),
		Y: float64(triangle.GetVertex2().GetY()),
		Z: float64(triangle.GetVertex2().GetZ()),
	}

	u0 := float64(triangle.GetUv0().GetU())
	v0 := float64(triangle.GetUv0().GetV())
	u1 := float64(triangle.GetUv1().GetU())
	v1 := float64(triangle.GetUv1().GetV())
	u2 := float64(triangle.GetUv2().GetU())
	v2 := float64(triangle.GetUv2().GetV())

	tri := hitable.NewTriangleWithUV(vertex0, vertex1, vertex2, u0, v0, u1, v1, u2, v2, material)

	return tri, nil
}

func (t *Transport) toSceneSpheres() ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}

	for _, sphere := range t.protoScene.GetObjects().GetSpheres() {
		sphere, err := t.toSceneSphere(sphere)
		if err != nil {
			return nil, err
		}
		hitables = append(hitables, sphere)
	}

	return hitables, nil
}

func (t *Transport) toSceneSphere(sphere *pb_transport.Sphere) (*hitable.Sphere, error) {
	material, ok := t.materials[sphere.GetMaterialName()]
	if !ok {
		return nil, fmt.Errorf("material %s not found", sphere.GetMaterialName())
	}

	center := &vec3.Vec3Impl{
		X: float64(sphere.GetCenter().GetX()),
		Y: float64(sphere.GetCenter().GetY()),
		Z: float64(sphere.GetCenter().GetZ()),
	}

	radius := float64(sphere.GetRadius())

	return hitable.NewSphere(center, center, 0, 1, radius, material), nil
}
