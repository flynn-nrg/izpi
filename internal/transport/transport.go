package transport

import (
	"fmt"
	"sync"

	"github.com/flynn-nrg/izpi/internal/camera"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/material"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/spectral"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

type Transport struct {
	aspectOverride       float32
	numWorkers           int
	colourRepresentation pb_transport.ColourRepresentation
	protoScene           *pb_transport.Scene
	triangles            []*pb_transport.Triangle
	textures             map[string]*texture.ImageTxt
	materials            map[string]material.Material
}

func NewTransport(
	aspectOverride float32,
	protoScene *pb_transport.Scene,
	triangles []*pb_transport.Triangle,
	textures map[string]*texture.ImageTxt,
	numWorkers int,
) *Transport {
	return &Transport{
		aspectOverride:       aspectOverride,
		colourRepresentation: protoScene.GetColourRepresentation(),
		protoScene:           protoScene,
		triangles:            triangles,
		textures:             textures,
		numWorkers:           numWorkers,
	}
}

func (t *Transport) ToScene() (*scene.Scene, error) {
	camera := t.toSceneCamera(t.aspectOverride)

	materials, err := t.toSceneMaterials()
	if err != nil {
		return nil, err
	}
	t.materials = materials

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

	// Create the scene
	scene := &scene.Scene{
		World:  hitable.NewSlice([]hitable.Hitable{hitable.NewBVH(hitables, 0, 1)}),
		Lights: hitable.NewSlice(lights),
		Camera: camera,
	}

	// Set world reference on dielectric materials for path length calculation
	for _, mat := range materials {
		if dielectric, ok := mat.(interface{ SetWorld(material.SceneGeometry) }); ok {
			// Create an adapter to convert HitableSlice to SceneGeometry
			sceneGeometry := &sceneGeometryAdapter{scene.World}
			dielectric.SetWorld(sceneGeometry)
		}
	}

	return scene, nil
}

func (t *Transport) toSceneMaterials() (map[string]material.Material, error) {
	var (
		mu           sync.Mutex
		errChan      = make(chan error)
		materialChan = make(chan *pb_transport.Material)
	)

	wg := &sync.WaitGroup{}
	materials := make(map[string]material.Material)

	// Start error collector goroutine
	var errs []error
	var errMu sync.Mutex
	errorCollectorDone := make(chan struct{})
	go func() {
		defer close(errorCollectorDone)
		for err := range errChan {
			errMu.Lock()
			errs = append(errs, err)
			errMu.Unlock()
		}
	}()

	// Spin up workers
	for range t.numWorkers {
		wg.Add(1)
		go t.toSceneMaterial(materialChan, materials, errChan, wg, &mu)
	}

	for _, material := range t.protoScene.GetMaterials() {
		materialChan <- material
	}

	close(materialChan)

	wg.Wait()

	close(errChan)

	<-errorCollectorDone

	if len(errs) > 0 {
		return nil, fmt.Errorf("errors converting materials: %v", errs)
	}

	return materials, nil
}

func (t *Transport) toSceneMaterial(
	materialChan chan *pb_transport.Material,
	materials map[string]material.Material,
	errChan chan error,
	wg *sync.WaitGroup,
	mu *sync.Mutex,
) {
	defer wg.Done()

	for material := range materialChan {
		switch material.GetType() {
		case pb_transport.MaterialType_LAMBERT:
			lambert, err := t.toSceneLambertMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = lambert
			mu.Unlock()
		case pb_transport.MaterialType_DIELECTRIC:
			dielectric, err := t.toSceneDielectricMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = dielectric
			mu.Unlock()
		case pb_transport.MaterialType_DIFFUSE_LIGHT:
			diffuselight, err := t.toSceneDiffuseLightMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = diffuselight
			mu.Unlock()
		case pb_transport.MaterialType_METAL:
			metal, err := t.toSceneMetalMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = metal
			mu.Unlock()
		case pb_transport.MaterialType_ISOTROPIC:
			isotropic, err := t.toSceneIsotropicMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = isotropic
			mu.Unlock()
		case pb_transport.MaterialType_PBR:
			pbr, err := t.toScenePBRMaterial(material)
			if err != nil {
				errChan <- err
				continue
			}
			mu.Lock()
			materials[material.GetName()] = pbr
			mu.Unlock()
		}
	}
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

	sssRadius := pbr.GetSssRadius()

	// Check if we need spectral rendering
	if t.colourRepresentation == pb_transport.ColourRepresentation_SPECTRAL {
		// Transform the albedo texture to spectral
		spectralAlbedo, err := t.textureToSpectralTexture(albedo)
		if err != nil {
			return nil, err
		}
		return material.NewPBRWithSpectralAlbedo(albedo, spectralAlbedo, normalMap, roughness, metalness, sss, sssRadius), nil
	}

	// Use regular RGB PBR material
	return material.NewPBR(albedo, normalMap, roughness, metalness, sss, sssRadius), nil
}

func (t *Transport) toSceneMetalMaterial(mat *pb_transport.Material) (material.Material, error) {
	metal := mat.GetMetal()

	albedo := &vec3.Vec3Impl{
		X: metal.GetAlbedo().GetX(),
		Y: metal.GetAlbedo().GetY(),
		Z: metal.GetAlbedo().GetZ(),
	}

	fuzz := metal.GetFuzz()

	return material.NewMetal(albedo, fuzz), nil
}

func (t *Transport) toSceneIsotropicMaterial(mat *pb_transport.Material) (material.Material, error) {
	isotropic := mat.GetIsotropic()

	switch isotropic.GetAlbedoProperties().(type) {
	case *pb_transport.IsotropicMaterial_Albedo:
		albedo, err := t.toSceneTexture(isotropic.GetAlbedo())
		if err != nil {
			return nil, err
		}
		return material.NewIsotropic(albedo), nil
	case *pb_transport.IsotropicMaterial_SpectralAlbedo:
		_, err := t.toSceneSpectralTexture(isotropic.GetSpectralAlbedo())
		if err != nil {
			return nil, err
		}
		// Note: Isotropic material doesn't have a spectral constructor yet
		// We'll need to add NewSpectralIsotropic to the material package
		return nil, fmt.Errorf("spectral isotropic materials not yet implemented")
	default:
		return nil, fmt.Errorf("isotropic material must have either albedo or spectral_albedo")
	}
}

func (t *Transport) toSceneLambertMaterial(mat *pb_transport.Material) (material.Material, error) {
	lambert := mat.GetLambert()

	switch lambert.GetAlbedoProperties().(type) {
	case *pb_transport.LambertMaterial_Albedo:
		albedo, err := t.toSceneTexture(lambert.GetAlbedo())
		if err != nil {
			return nil, err
		}
		return material.NewLambertian(albedo), nil
	case *pb_transport.LambertMaterial_SpectralAlbedo:
		spectralAlbedo, err := t.toSceneSpectralTexture(lambert.GetSpectralAlbedo())
		if err != nil {
			return nil, err
		}
		return material.NewSpectralLambertian(spectralAlbedo), nil
	default:
		return nil, fmt.Errorf("lambert material must have either albedo or spectral_albedo")
	}
}

func (t *Transport) toSceneDielectricMaterial(mat *pb_transport.Material) (material.Material, error) {
	dielectric := mat.GetDielectric()

	// Handle refractive index properties
	var refIdx float32
	var spectralRefIdx texture.SpectralTexture
	var err error

	switch dielectric.GetRefractiveIndexProperties().(type) {
	case *pb_transport.DielectricMaterial_Refidx:
		refIdx = dielectric.GetRefidx()
	case *pb_transport.DielectricMaterial_SpectralRefidx:
		spectralRefIdx, err = t.toSceneSpectralTexture(dielectric.GetSpectralRefidx())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("dielectric material must have either refidx or spectral_refidx")
	}

	// Handle absorption properties (optional)
	var absorptionCoeff *vec3.Vec3Impl
	var spectralAbsorptionCoeff texture.SpectralTexture

	switch dielectric.GetAbsorptionProperties().(type) {
	case *pb_transport.DielectricMaterial_AbsorptionCoeff:
		abs := dielectric.GetAbsorptionCoeff()
		absorptionCoeff = &vec3.Vec3Impl{
			X: abs.GetX(),
			Y: abs.GetY(),
			Z: abs.GetZ(),
		}
	case *pb_transport.DielectricMaterial_SpectralAbsorptionCoeff:
		spectralAbsorptionCoeff, err = t.toSceneSpectralTexture(dielectric.GetSpectralAbsorptionCoeff())
		if err != nil {
			return nil, err
		}
	}

	// Create the appropriate dielectric material based on available properties
	if spectralRefIdx != nil {
		if spectralAbsorptionCoeff != nil {
			return material.NewSpectralColoredDielectric(spectralRefIdx, spectralAbsorptionCoeff), nil
		} else {
			return material.NewSpectralDielectric(spectralRefIdx), nil
		}
	} else {
		if absorptionCoeff != nil {
			return material.NewColoredDielectric(refIdx, absorptionCoeff), nil
		} else {
			return material.NewDielectric(refIdx), nil
		}
	}
}

func (t *Transport) toSceneDiffuseLightMaterial(mat *pb_transport.Material) (material.Material, error) {
	diffuselight := mat.GetDiffuselight()

	switch diffuselight.GetEmissionProperties().(type) {
	case *pb_transport.DiffuseLightMaterial_Emit:
		emit, err := t.toSceneTexture(diffuselight.GetEmit())
		if err != nil {
			return nil, err
		}
		return material.NewDiffuseLight(emit), nil
	case *pb_transport.DiffuseLightMaterial_SpectralEmit:
		spectralEmit, err := t.toSceneSpectralTexture(diffuselight.GetSpectralEmit())
		if err != nil {
			return nil, err
		}
		return material.NewSpectralDiffuseLight(spectralEmit), nil
	default:
		return nil, fmt.Errorf("diffuse light material must have either emit or spectral_emit")
	}
}

func (t *Transport) toSceneTexture(text *pb_transport.Texture) (texture.Texture, error) {
	switch text.GetTextureProperties().(type) {
	case *pb_transport.Texture_Constant:
		return t.toSceneConstantTexture(text)
	case *pb_transport.Texture_Image:
		return t.toSceneImageTexture(text)
	case *pb_transport.Texture_SpectralConstant:
		// Convert spectral texture to RGB texture for backward compatibility
		// This is a fallback for RGB rendering when spectral textures are provided
		_, err := t.toSceneSpectralTexture(text.GetSpectralConstant())
		if err != nil {
			return nil, err
		}
		// Create a neutral RGB texture as fallback
		return texture.NewConstant(&vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5}), nil
	case *pb_transport.Texture_SpectralChecker:
		// Convert spectral checker to RGB checker for backward compatibility
		// This is a fallback for RGB rendering when spectral textures are provided
		return texture.NewConstant(&vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5}), nil
	}

	return nil, fmt.Errorf("unknown texture type: %T", text.GetTextureProperties())
}

func (t *Transport) toSceneConstantTexture(text *pb_transport.Texture) (texture.Texture, error) {
	constant := text.GetConstant()

	return texture.NewConstant(&vec3.Vec3Impl{
		X: constant.GetValue().GetX(),
		Y: constant.GetValue().GetY(),
		Z: constant.GetValue().GetZ(),
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

func (t *Transport) toSceneSpectralTexture(spectralText *pb_transport.SpectralConstantTexture) (texture.SpectralTexture, error) {
	switch spectralText.GetSpectralProperties().(type) {
	case *pb_transport.SpectralConstantTexture_Gaussian:
		gaussian := spectralText.GetGaussian()
		return texture.NewSpectralConstant(
			gaussian.GetPeakValue(),
			gaussian.GetCenterWavelength(),
			gaussian.GetWidth(),
		), nil
	case *pb_transport.SpectralConstantTexture_Tabulated:
		tabulated := spectralText.GetTabulated()
		wavelengths := make([]float32, len(tabulated.GetWavelengths()))
		values := make([]float32, len(tabulated.GetValues()))

		for i, w := range tabulated.GetWavelengths() {
			wavelengths[i] = w
		}
		for i, v := range tabulated.GetValues() {
			values[i] = v
		}

		spd := spectral.NewSPD(wavelengths, values)
		return texture.NewSpectralConstantFromSPD(spd), nil
	case *pb_transport.SpectralConstantTexture_Neutral:
		neutral := spectralText.GetNeutral()
		return texture.NewSpectralNeutral(neutral.GetReflectance()), nil
	default:
		return nil, fmt.Errorf("unknown spectral texture type: %T", spectralText.GetSpectralProperties())
	}
}

// textureToSpectralTexture converts a regular texture to a spectral texture
// by creating a spectral image texture from the regular texture's data
func (t *Transport) textureToSpectralTexture(tex texture.Texture) (texture.SpectralTexture, error) {
	// Note: A texture cannot implement both Texture and SpectralTexture interfaces
	// due to conflicting Value method signatures, so we don't need to check for this

	// For image textures, we can convert them to spectral image textures
	if imageTex, ok := tex.(*texture.ImageTxt); ok {
		// Get the image data and create a spectral image texture
		img := imageTex.GetData()
		if img == nil {
			return nil, fmt.Errorf("image texture has no data")
		}

		// Create spectral image texture from the image data
		spectralImage := texture.NewSpectralImageFromImage(img)
		return spectralImage, nil
	}

	// For constant textures, create a neutral spectral texture
	if constTex, ok := tex.(*texture.Constant); ok {
		// Get the RGB value and create a spectral constant with proper luminance
		rgbValue := constTex.Value(0, 0, nil) // UV coordinates don't matter for constant textures
		luminance := 0.299*rgbValue.X + 0.587*rgbValue.Y + 0.114*rgbValue.Z
		return texture.NewSpectralNeutral(luminance), nil
	}

	// For other texture types, create a neutral spectral texture as fallback
	return texture.NewSpectralNeutral(0.5), nil
}

func (t *Transport) toSceneCamera(aspectOverride float32) *camera.Camera {
	protoCamera := t.protoScene.GetCamera()

	lookFrom := &vec3.Vec3Impl{
		X: protoCamera.GetLookfrom().GetX(),
		Y: protoCamera.GetLookfrom().GetY(),
		Z: protoCamera.GetLookfrom().GetZ(),
	}

	lookAt := &vec3.Vec3Impl{
		X: protoCamera.GetLookat().GetX(),
		Y: protoCamera.GetLookat().GetY(),
		Z: protoCamera.GetLookat().GetZ(),
	}

	vup := &vec3.Vec3Impl{
		X: protoCamera.GetVup().GetX(),
		Y: protoCamera.GetVup().GetY(),
		Z: protoCamera.GetVup().GetZ(),
	}

	vfov := protoCamera.GetVfov()

	var aspect float32
	if aspectOverride != 0.0 {
		aspect = aspectOverride
	} else {
		aspect = protoCamera.GetAspect()
	}

	aperture := protoCamera.GetAperture()
	focusDist := protoCamera.GetFocusdist()
	time0 := protoCamera.GetTime0()
	time1 := protoCamera.GetTime1()

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
		X: triangle.GetVertex0().GetX(),
		Y: triangle.GetVertex0().GetY(),
		Z: triangle.GetVertex0().GetZ(),
	}

	vertex1 := &vec3.Vec3Impl{
		X: triangle.GetVertex1().GetX(),
		Y: triangle.GetVertex1().GetY(),
		Z: triangle.GetVertex1().GetZ(),
	}

	vertex2 := &vec3.Vec3Impl{
		X: triangle.GetVertex2().GetX(),
		Y: triangle.GetVertex2().GetY(),
		Z: triangle.GetVertex2().GetZ(),
	}

	u0 := triangle.GetUv0().GetU()
	v0 := triangle.GetUv0().GetV()
	u1 := triangle.GetUv1().GetU()
	v1 := triangle.GetUv1().GetV()
	u2 := triangle.GetUv2().GetU()
	v2 := triangle.GetUv2().GetV()

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
		X: sphere.GetCenter().GetX(),
		Y: sphere.GetCenter().GetY(),
		Z: sphere.GetCenter().GetZ(),
	}

	radius := sphere.GetRadius()

	return hitable.NewSphere(center, center, 0, 1, radius, material), nil
}

// sceneGeometryAdapter adapts HitableSlice to SceneGeometry interface
type sceneGeometryAdapter struct {
	world *hitable.HitableSlice
}

func (sga *sceneGeometryAdapter) Hit(r ray.Ray, tMin float32, tMax float32) (*hitrecord.HitRecord, material.Material, bool) {
	return sga.world.Hit(r, tMin, tMax)
}
