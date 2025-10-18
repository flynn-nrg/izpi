// Package scenes implements some sample scenes.
package scenes

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flynn-nrg/izpi/internal/camera"
	"github.com/flynn-nrg/izpi/internal/displacement"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/flynn-nrg/izpi/internal/wavefront"
	"google.golang.org/protobuf/proto"

	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	log "github.com/sirupsen/logrus"
)

// RandomScene returns a random scene.
func RandomScene() *hitable.HitableSlice {
	checker := texture.NewChecker(texture.NewConstant(&vec3.Vec3Impl{X: 0.2, Y: 0.3, Z: 0.1}),
		texture.NewConstant(&vec3.Vec3Impl{X: 0.9, Y: 0.9, Z: 0.9}))
	spheres := []hitable.Hitable{hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: -1000, Z: 0}, &vec3.Vec3Impl{X: 0, Y: -1000, Z: 0}, 0, 1, 1000, material.NewLambertian(checker))}
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rand.float32()
			center := &vec3.Vec3Impl{X: float32(a) + 0.9*rand.float32(), Y: 0.2, Z: float32(b) + 0.9*rand.float32()}
			if vec3.Sub(center, &vec3.Vec3Impl{X: 4, Y: 0.2, Z: 0}).Length() > 0.9 {
				if chooseMat < 0.8 {
					// diffuse
					spheres = append(spheres, hitable.NewSphere(center,
						vec3.Add(center, &vec3.Vec3Impl{Y: 0.5 * rand.float32()}), 0.0, 1.0, 0.2,
						material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{
							X: rand.float32() * rand.float32(),
							Y: rand.float32() * rand.float32(),
							Z: rand.float32() * rand.float32(),
						}))))
				} else if chooseMat < 0.95 {
					// metal
					spheres = append(spheres, hitable.NewSphere(center, center, 0.0, 1.0, 0.2,
						material.NewMetal(&vec3.Vec3Impl{
							X: 0.5 * (1.0 - rand.float32()),
							Y: 0.5 * (1.0 - rand.float32()),
							Z: 0.5 * (1.0 - rand.float32()),
						}, 0.2*rand.float32())))
				} else {
					// glass
					spheres = append(spheres, hitable.NewSphere(center, center, 0.0, 1.0, 0.2, material.NewDielectric(1.5)))
				}
			}
		}
	}

	spheres = append(spheres, hitable.NewSphere(&vec3.Vec3Impl{Y: 1.0}, &vec3.Vec3Impl{Y: 1.0}, 0.0, 1.0, 1.0, material.NewDielectric(1.5)))
	spheres = append(spheres, hitable.NewSphere(&vec3.Vec3Impl{X: -4.0, Y: 1.0}, &vec3.Vec3Impl{X: -4.0, Y: 1.0}, 0.0, 1.0, 1.0, material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.4, Y: 0.2, Z: 0.1}))))
	spheres = append(spheres, hitable.NewSphere(&vec3.Vec3Impl{X: 4.0, Y: 1.0}, &vec3.Vec3Impl{X: 4.0, Y: 1.0}, 0.0, 1.0, 1.0, material.NewMetal(&vec3.Vec3Impl{X: 0.7, Y: 0.6, Z: 0.5}, 0.0)))

	return hitable.NewSlice(spheres)
}

// TwoSpheres returns a scene containing two spheres.
func TwoSpheres() *hitable.HitableSlice {
	checker := texture.NewChecker(texture.NewConstant(&vec3.Vec3Impl{X: 0.2, Y: 0.3, Z: 0.1}),
		texture.NewConstant(&vec3.Vec3Impl{X: 0.9, Y: 0.9, Z: 0.9}))
	spheres := []hitable.Hitable{
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: -10, Z: 0}, &vec3.Vec3Impl{X: 0, Y: -10, Z: 0}, 0, 1, 10, material.NewLambertian(checker)),
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 10, Z: 0}, &vec3.Vec3Impl{X: 0, Y: 10, Z: 0}, 0, 1, 10, material.NewLambertian(checker)),
	}

	return hitable.NewSlice(spheres)
}

// TwoPerlinSpheres returns a scene containing two spheres with Perlin noise.
func TwoPerlinSpheres() *hitable.HitableSlice {
	perText := texture.NewNoise(4.0)
	spheres := []hitable.Hitable{
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: -1000, Z: 0}, &vec3.Vec3Impl{X: 0, Y: -1000, Z: 0}, 0, 1, 1000, material.NewLambertian(perText)),
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 2, Z: 0}, &vec3.Vec3Impl{X: 0, Y: 2, Z: 0}, 0, 1, 2, material.NewLambertian(perText)),
	}

	return hitable.NewSlice(spheres)
}

// TextureMappedSphere returns a scene containing a representation of Earth.
func TextureMappedSphere() *hitable.HitableSlice {
	file, err := os.Open("../images/earth.png")
	if err != nil {
		log.Fatalf("could not read texture file; %v", err)
	}
	imgText, err := texture.NewFromPNG(file)
	if err != nil {
		log.Fatalf("failed to decode image; %v", err)
	}
	spheres := []hitable.Hitable{
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 0, Z: 0}, &vec3.Vec3Impl{X: 0, Y: 0, Z: 0}, 0, 1, 1, material.NewLambertian(imgText)),
	}

	return hitable.NewSlice(spheres)
}

// SimpleLight returns a scene containing three spheres and a rectangle.
func SimpleLight() *hitable.HitableSlice {
	perText := texture.NewNoise(4.0)
	hitables := []hitable.Hitable{
		hitable.NewSphere(&vec3.Vec3Impl{Y: -1000}, &vec3.Vec3Impl{Y: -1000}, 0, 1, 1000, material.NewLambertian(perText)),
		hitable.NewSphere(&vec3.Vec3Impl{Y: 2}, &vec3.Vec3Impl{Y: 2}, 0, 1, 2, material.NewLambertian(perText)),
		hitable.NewSphere(&vec3.Vec3Impl{Y: 7}, &vec3.Vec3Impl{Y: 7}, 0, 1, 2, material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 4, Y: 4, Z: 4}))),
		hitable.NewXYRect(3, 5, 1, 3, -2, material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 4, Y: 4, Z: 4}))),
	}

	return hitable.NewSlice(hitables)
}

// CornellBox returns a scene recreating the Cornell box.
func CornellBox(aspect float32) *scene.Scene {
	red := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.65, Y: 0.05, Z: 0.05}))
	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	green := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.12, Y: 0.45, Z: 0.15}))
	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 15, Y: 15, Z: 15}))
	glass := material.NewDielectric(1.5)
	hitables := []hitable.Hitable{
		hitable.NewFlipNormals(hitable.NewYZRect(0, 555, 0, 555, 555, green)),
		hitable.NewYZRect(0, 555, 0, 555, 0, red),
		hitable.NewFlipNormals(hitable.NewXZRect(213, 343, 227, 332, 554, light)),
		hitable.NewFlipNormals(hitable.NewXZRect(0, 555, 0, 555, 555, white)),
		hitable.NewXZRect(0, 555, 0, 555, 0, white),
		hitable.NewFlipNormals(hitable.NewXYRect(0, 555, 0, 555, 555, white)),
		hitable.NewSphere(&vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, &vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, 0, 1, 90, glass),
		hitable.NewTranslate(hitable.NewRotateY(hitable.NewBox(&vec3.Vec3Impl{X: 0, Y: 0, Z: 0}, &vec3.Vec3Impl{X: 165, Y: 330, Z: 165}, white), 15), &vec3.Vec3Impl{X: 265, Y: 0, Z: 295}),
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 278.0, Y: 278.0, Z: -800.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam)
}

// Final returns the scene from the last chapter in the book.
func Final(aspect float32) (*hitable.HitableSlice, *camera.Camera) {
	nb := 20
	list := []hitable.Hitable{}
	boxList := []hitable.Hitable{}
	boxList2 := []hitable.Hitable{}

	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	ground := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.48, Y: 0.83, Z: 0.53}))

	for i := 0; i < nb; i++ {
		for j := 0; j < nb; j++ {
			w := float32(100)
			x0 := -1000.0 + float32(i)*w
			z0 := -1000.0 + float32(j)*w
			y0 := float32(0)
			x1 := x0 + w
			y1 := 100.0 * (rand.float32() + 0.01)
			z1 := z0 + w
			boxList = append(boxList, hitable.NewBox(&vec3.Vec3Impl{X: x0, Y: y0, Z: z0}, &vec3.Vec3Impl{X: x1, Y: y1, Z: z1}, ground))
		}
	}

	list = append(list, hitable.NewBVH(boxList, 0, 1))

	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 7, Y: 7, Z: 7}))
	list = append(list, hitable.NewXZRect(123, 423, 147, 412, 554, light))

	center := &vec3.Vec3Impl{X: 400, Y: 400, Z: 350}
	list = append(list, hitable.NewSphere(center, vec3.Add(center, &vec3.Vec3Impl{X: 30}), 0, 1, 50, material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.7, Y: 0.3, Z: 0.1}))))
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 260, Y: 150, Z: 45}, &vec3.Vec3Impl{X: 260, Y: 150, Z: 45}, 0, 1, 50, material.NewDielectric(1.5)))
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 150, Z: 145}, &vec3.Vec3Impl{X: 0, Y: 150, Z: 145}, 0, 1, 50, material.NewMetal(&vec3.Vec3Impl{X: 0.8, Y: 0.8, Z: 0.9}, 10.0)))

	boundary := hitable.NewSphere(&vec3.Vec3Impl{X: 360, Y: 150, Z: 145}, &vec3.Vec3Impl{X: 360, Y: 150, Z: 145}, 0, 1, 70, material.NewDielectric(1.5))
	list = append(list, boundary)
	list = append(list, hitable.NewConstantMedium(boundary, 0.2, texture.NewConstant(&vec3.Vec3Impl{X: 0.2, Y: 0.4, Z: 0.9})))
	boundary = hitable.NewSphere(&vec3.Vec3Impl{}, &vec3.Vec3Impl{}, 0, 1, 5000, material.NewDielectric(1.5))
	list = append(list, hitable.NewConstantMedium(boundary, 0.0001, texture.NewConstant(&vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0})))

	file, err := os.Open("../images/earth.png")
	if err != nil {
		log.Fatalf("could not read texture file; %v", err)
	}
	imgText, err := texture.NewFromPNG(file)
	if err != nil {
		log.Fatalf("failed to decode image; %v", err)
	}
	emat := material.NewLambertian(imgText)
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 400, Y: 300, Z: 400}, &vec3.Vec3Impl{X: 400, Y: 300, Z: 400}, 0, 1, 100, emat))

	perText := texture.NewNoise(0.1)
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 220, Y: 280, Z: 300}, &vec3.Vec3Impl{X: 220, Y: 280, Z: 300}, 0, 1, 80, material.NewLambertian(perText)))

	ns := 1000
	for j := 0; j < ns; j++ {
		center := &vec3.Vec3Impl{X: 165 * rand.float32(), Y: 165 * rand.float32(), Z: 165 * rand.float32()}
		boxList2 = append(boxList2, hitable.NewSphere(center, center, 0, 1, 10, white))
	}

	list = append(list, hitable.NewTranslate(hitable.NewRotateY(hitable.NewBVH(boxList2, 0, 1), 15), &vec3.Vec3Impl{X: -100, Y: 270, Z: 395}))

	lookFrom := &vec3.Vec3Impl{X: 478.0, Y: 278.0, Z: -600.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return hitable.NewSlice(list), cam
}

// Environment returns a scene that tests the sky dome and HDR textures functionality.
func Environment(aspect float32) *scene.Scene {
	dome, err := hitable.NewSkyDome(&vec3.Vec3Impl{}, 100, "decor_shop_4k.hdr")
	if err != nil {
		log.Fatal(err)
	}

	glass := material.NewDielectric(1.5)
	glassSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, &vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, 0, 1, 4, glass)
	metal := material.NewMetal(&vec3.Vec3Impl{X: 0.5, Y: 1.0, Z: 1.0}, 0)
	metalSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, &vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, 0, 1, 3, metal)
	hitables := []hitable.Hitable{glassSphere, metalSphere, dome}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 0, Y: 0, Z: 10}
	lookAt := &vec3.Vec3Impl{X: -20, Y: 0, Z: -1}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(60.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam)

}

// CornellBox returns a scene recreating the Cornell box.
func CornellBoxObj(aspect float32) (*scene.Scene, error) {
	red := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.65, Y: 0.05, Z: 0.05}))
	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	green := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.12, Y: 0.45, Z: 0.15}))
	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 15, Y: 15, Z: 15}))
	gold := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: .7, Y: .7, Z: .85}))
	glass := material.NewDielectric(1.5)

	/*
		plaFile, err := os.Open("platano/ripe-banana_u1_v1.png")
		if err != nil {
			return nil, err
		}
		plaTex, err := texture.NewFromPNG(plaFile)
		if err != nil {
			return nil, err
		}
		platano := material.NewLambertian(plaTex)
	*/
	objectName := "PP.obj"
	r, err := os.Open(objectName)
	if err != nil {
		return nil, err
	}

	cube, err := wavefront.NewObjFromReader(r, filepath.Dir(objectName))
	if err != nil {
		return nil, err
	}

	cube.Translate(&vec3.Vec3Impl{X: 280, Y: 30, Z: 390})
	cube.Scale(&vec3.Vec3Impl{X: 14, Y: 14, Z: 14})

	hitables := []hitable.Hitable{
		hitable.NewFlipNormals(hitable.NewYZRect(0, 555, 0, 555, 555, green)),
		hitable.NewYZRect(0, 555, 0, 555, 0, red),
		hitable.NewFlipNormals(hitable.NewXZRect(213, 343, 227, 332, 554, light)),
		hitable.NewFlipNormals(hitable.NewXZRect(0, 555, 0, 555, 555, white)),
		hitable.NewXZRect(0, 555, 0, 555, 0, white),
		hitable.NewFlipNormals(hitable.NewXYRect(0, 555, 0, 555, 555, white)),
		hitable.NewSphere(&vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, &vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, 0, 1, 90, glass),
	}

	for i := range cube.Groups {
		cubeHitables, err := cube.GroupToHitablesWithCustomMaterial(i, gold)
		if err != nil {
			return nil, err
		}
		bvh := hitable.NewBVH(cubeHitables, 0, 1)
		hitables = append(hitables, bvh)
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 278.0, Y: 278.0, Z: -800.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

func TVSet(aspect float32) (*scene.Scene, error) {
	red := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.65, Y: 0.05, Z: 0.05}))
	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	tvTextFile, err := os.Open("Television_01_diff_4k.png")
	if err != nil {
		return nil, err
	}
	tvText, err := texture.NewFromPNG(tvTextFile)
	if err != nil {
		return nil, err
	}
	tvMat := material.NewLambertian(tvText)
	green := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.12, Y: 0.45, Z: 0.15}))
	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 15, Y: 15, Z: 15}))
	glass := material.NewDielectric(1.5)

	objectName := "Television_01_4k.obj"
	r, err := os.Open(objectName)
	if err != nil {
		return nil, err
	}
	cube, err := wavefront.NewObjFromReader(r, filepath.Dir(objectName))
	if err != nil {
		return nil, err
	}

	cube.Scale(&vec3.Vec3Impl{X: 500, Y: 500, Z: 500})
	cube.Translate(&vec3.Vec3Impl{X: 280, Y: 100, Z: 450})

	hitables := []hitable.Hitable{
		hitable.NewFlipNormals(hitable.NewYZRect(0, 555, 0, 555, 555, green)),
		hitable.NewYZRect(0, 555, 0, 555, 0, red),
		hitable.NewFlipNormals(hitable.NewXZRect(213, 343, 227, 332, 554, light)),
		hitable.NewFlipNormals(hitable.NewXZRect(0, 555, 0, 555, 555, white)),
		hitable.NewXZRect(0, 555, 0, 555, 0, white),
		hitable.NewFlipNormals(hitable.NewXYRect(0, 555, 0, 555, 555, white)),
		hitable.NewSphere(&vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, &vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, 0, 1, 90, glass),
	}

	for i := range cube.Groups {
		fmt.Printf("Grupana...\n")
		cubeHitables, err := cube.GroupToHitablesWithCustomMaterial(i, tvMat)
		if err != nil {
			return nil, err
		}
		bvh := hitable.NewBVH(cubeHitables, 0, 1)
		hitables = append(hitables, bvh)
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 278.0, Y: 278.0, Z: -800.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

func SWHangar(aspect float32) (*scene.Scene, error) {
	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	glass := material.NewDielectric(1.5)

	objectName := "sw/hangar.obj"
	objFile, err := os.Open(objectName)
	if err != nil {
		return nil, err
	}

	hangar, err := wavefront.NewObjFromReader(objFile, filepath.Dir(objectName), wavefront.IGNORE_TEXTURES)
	if err != nil {
		return nil, err
	}

	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 10, Y: 10, Z: 10}))

	hitables := []hitable.Hitable{
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 30, Z: -30}, &vec3.Vec3Impl{X: 0, Y: 30, Z: -30}, 0, 1, 20, light),
		hitable.NewSphere(&vec3.Vec3Impl{X: -50, Y: 15, Z: 65}, &vec3.Vec3Impl{X: -50, Y: 15, Z: 65}, 0, 1, 20, glass),
	}

	for i := range hangar.Groups {
		hangarHitables, err := hangar.GroupToHitablesWithCustomMaterial(i, white)
		if err != nil {
			return nil, err
		}
		bvh := hitable.NewBVH(hangarHitables, 0, 1)
		hitables = append(hitables, bvh)
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 0.0, Y: 33.0, Z: 30.0}
	lookAt := &vec3.Vec3Impl{X: -1, Y: 33, Z: 31}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(100.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

// DisplacementTest returns a scene recreating the Cornell box and a displacement map applied to the floor.
func DisplacementTest(aspect float32) (*scene.Scene, error) {
	red := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.65, Y: 0.05, Z: 0.05}))
	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	green := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.12, Y: 0.45, Z: 0.15}))
	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 15, Y: 15, Z: 15}))
	glass := material.NewDielectric(1.5)

	// https://ambientcg.com/view?id=Bricks078
	floorTextFile, err := os.Open("bricks/Bricks078_4K_Color.png")
	if err != nil {
		return nil, err
	}

	floorText, err := texture.NewFromPNG(floorTextFile)
	if err != nil {
		return nil, err
	}

	floorMat := material.NewLambertian(floorText)

	fmt.Printf("%v", floorMat)

	floor := []*hitable.Triangle{
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 555}, &vec3.Vec3Impl{}, &vec3.Vec3Impl{X: 555, Z: 555}, 1, 0, 0, 0, 1, 1, floorMat),
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{}, &vec3.Vec3Impl{Z: 555}, &vec3.Vec3Impl{X: 555, Z: 555}, 0, 0, 0, 1, 1, 1, floorMat),
	}

	// https://ambientcg.com/view?id=Bricks078 scaled down to 400x200.
	displacementFile, err := os.Open("bricks/displacement.png")
	if err != nil {
		return nil, err
	}

	displacementText, err := texture.NewFromPNG(displacementFile)
	if err != nil {
		return nil, err
	}

	displacedFloor, err := displacement.ApplyDisplacementMap(floor, displacementText, 0, 20)
	if err != nil {
		return nil, err
	}

	hitables := []hitable.Hitable{
		hitable.NewFlipNormals(hitable.NewYZRect(0, 555, 0, 555, 555, green)),
		hitable.NewYZRect(0, 555, 0, 555, 0, red),
		hitable.NewFlipNormals(hitable.NewXZRect(213, 343, 227, 332, 554, light)),
		hitable.NewFlipNormals(hitable.NewXZRect(0, 555, 0, 555, 555, white)),
		hitable.NewFlipNormals(hitable.NewXYRect(0, 555, 0, 555, 555, white)),
		//	hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 555}, &vec3.Vec3Impl{}, &vec3.Vec3Impl{X: 555, Z: 555}, 1, 0, 0, 0, 1, 1, floorMat),
		//	hitable.NewTriangleWithUV(&vec3.Vec3Impl{}, &vec3.Vec3Impl{Z: 555}, &vec3.Vec3Impl{X: 555, Z: 555}, 0, 0, 0, 1, 1, 1, floorMat),
		hitable.NewSphere(&vec3.Vec3Impl{X: 190, Y: 130, Z: 190}, &vec3.Vec3Impl{X: 190, Y: 130, Z: 190}, 0, 1, 90, glass),
	}

	triangles := []hitable.Hitable{}
	for _, t := range displacedFloor {
		triangles = append(triangles, t)
	}
	bvh := hitable.NewBVH(triangles, 0, 1)

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	hitables = append(hitables, bvh)

	lookFrom := &vec3.Vec3Impl{X: 278.0, Y: 278.0, Z: -800.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

// VWBeetle returns a scene that tests the sky dome and HDR textures functionality.
func VWBeetle(aspect float32) (*scene.Scene, error) {
	dome, err := hitable.NewSkyDome(&vec3.Vec3Impl{}, 100, "OutdoorHDRI019_4K-HDR.hdr")
	if err != nil {
		return nil, err
	}

	// https://github.com/alecjacobson/common-3d-test-models
	beetleFile, err := os.Open("beetle.obj")
	if err != nil {
		return nil, err
	}

	beetleObj, err := wavefront.NewObjFromReader(beetleFile, "./", wavefront.IGNORE_MATERIALS)
	if err != nil {
		return nil, err
	}

	beetleObj.Scale(&vec3.Vec3Impl{X: 4, Y: 4, Z: 4})
	//	glass := material.NewDielectric(1.5)
	//	glassSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, &vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, 0, 1, 4, glass)
	//	metal := material.NewMetal(&vec3.Vec3Impl{X: 0.5, Y: 1.0, Z: 1.0}, 0)
	//	metalSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, &vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, 0, 1, 3, metal)
	hitables := []hitable.Hitable{dome}
	green := material.NewMetal(&vec3.Vec3Impl{X: 0.8, Y: .8, Z: .8}, 0.3)

	for i := range beetleObj.Groups {
		bettleHitables, err := beetleObj.GroupToHitablesWithCustomMaterial(i, green)
		if err != nil {
			return nil, err
		}
		bvh := hitable.NewBVH(bettleHitables, 0, 1)
		hitables = append(hitables, bvh)
	}

	lights := []hitable.Hitable{}
	for _, h := range hitables {
		if h.IsEmitter() {
			lights = append(lights, h)
		}
	}

	lookFrom := &vec3.Vec3Impl{X: 1.0, Y: 2, Z: 3.0}
	lookAt := &vec3.Vec3Impl{X: -.4, Y: 1.3, Z: 1}
	//lookFrom := &vec3.Vec3Impl{X: 2.0, Y: 0.6, Z: 2.0}
	//lookAt := &vec3.Vec3Impl{X: .4, Y: 0, Z: 1}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float32(60.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil

}

// Challenger returns a scene that tests the sky dome and HDR textures functionality.
func Challenger(aspect float32) ([]byte, error) {
	// https://www.cgtrader.com/free-3d-models/car/sport-car/dodge-challenger-87e47a62-3aaf-4d8f-84c9-6af70b9792b0

	challengerFile, err := os.Open("challenger_tri.obj")
	if err != nil {
		return nil, err
	}

	challengerObj, err := wavefront.NewObjFromReader(challengerFile, "./", wavefront.IGNORE_MATERIALS, wavefront.IGNORE_TEXTURES)
	if err != nil {
		return nil, err
	}

	challengerObj.Scale(&vec3.Vec3Impl{X: 6, Y: 6, Z: 6})
	challengerObj.Translate(&vec3.Vec3Impl{X: 0, Y: .47, Z: 0})

	carTriangles, err := challengerObj.GroupToTransportTrianglesWithMaterial(0, "car")
	if err != nil {
		return nil, err
	}

	/*
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: 50, Y: 0, Z: 50}, 1, 0, 0, 0, 1, 1, white),
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: 50}, &vec3.Vec3Impl{X: 50, Y: 0, Z: 50}, 0, 0, 0, 1, 1, 1, white),
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: 50, Y: 50, Z: 50}, 1, 0, 0, 0, 1, 1, metal),
		hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -50, Y: -5, Z: -50}, &vec3.Vec3Impl{X: -50, Y: -5, Z: 50}, &vec3.Vec3Impl{X: -50, Y: 85, Z: -50}, 1, 0, 0, 0, 1, 1, metal),
	*/
	sceneTriangles := []*pb_transport.Triangle{
		{
			Vertex0:      &pb_transport.Vec3{X: 50, Y: 0, Z: -50},
			Vertex1:      &pb_transport.Vec3{X: -50, Y: 0, Z: -50},
			Vertex2:      &pb_transport.Vec3{X: 50, Y: 0, Z: 50},
			Uv0:          &pb_transport.Vec2{U: 1, V: 0},
			Uv1:          &pb_transport.Vec2{U: 0, V: 0},
			Uv2:          &pb_transport.Vec2{U: 1, V: 1},
			MaterialName: "white",
		},
		{
			Vertex0:      &pb_transport.Vec3{X: -50, Y: 0, Z: -50},
			Vertex1:      &pb_transport.Vec3{X: -50, Y: 0, Z: 50},
			Vertex2:      &pb_transport.Vec3{X: 50, Y: 0, Z: 50},
			Uv0:          &pb_transport.Vec2{U: 0, V: 0},
			Uv1:          &pb_transport.Vec2{U: 0, V: 1},
			Uv2:          &pb_transport.Vec2{U: 1, V: 1},
			MaterialName: "white",
		},
		{
			Vertex0:      &pb_transport.Vec3{X: 50, Y: 0, Z: -50},
			Vertex1:      &pb_transport.Vec3{X: -50, Y: 0, Z: -50},
			Vertex2:      &pb_transport.Vec3{X: 50, Y: 50, Z: 50},
			Uv0:          &pb_transport.Vec2{U: 1, V: 0},
			Uv1:          &pb_transport.Vec2{U: 0, V: 0},
			Uv2:          &pb_transport.Vec2{U: 1, V: 1},
			MaterialName: "metal",
		},
		{
			Vertex0:      &pb_transport.Vec3{X: -50, Y: -5, Z: -50},
			Vertex1:      &pb_transport.Vec3{X: -50, Y: -5, Z: 50},
			Vertex2:      &pb_transport.Vec3{X: -50, Y: 85, Z: -50},
			Uv0:          &pb_transport.Vec2{U: 1, V: 0},
			Uv1:          &pb_transport.Vec2{U: 0, V: 0},
			Uv2:          &pb_transport.Vec2{U: 1, V: 1},
			MaterialName: "metal",
		},
	}

	sceneTriangles = append(sceneTriangles, carTriangles...)

	protoScene := &pb_transport.Scene{
		Name:    "Challenger",
		Version: "1.0",
		Camera: &pb_transport.Camera{
			Lookfrom:  &pb_transport.Vec3{X: 18.0, Y: 8, Z: 12.0},
			Lookat:    &pb_transport.Vec3{X: -6, Y: .7, Z: 0},
			Vup:       &pb_transport.Vec3{X: 0, Y: 1, Z: 0},
			Vfov:      75.0,
			Aspect:    float32(aspect),
			Aperture:  0.1,
			Focusdist: float32(10.0),
			Time0:     0.0,
			Time1:     1.0,
		},
		Materials: map[string]*pb_transport.Material{
			"white": {
				Name: "white",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{X: 0.73, Y: 0.73, Z: 0.73},
									},
								},
							},
						},
					},
				},
			},
			"car": {
				Name: "car",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_roughness.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_normal-ogl.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
					},
				},
			},
			"light": {
				Name: "light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_Emit{
							Emit: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{X: 15, Y: 15, Z: 15},
									},
								},
							},
						},
					},
				},
			},
			"metal": {
				Name: "metal",
				Type: pb_transport.MaterialType_METAL,
				MaterialProperties: &pb_transport.Material_Metal{
					Metal: &pb_transport.MetalMaterial{
						Albedo: &pb_transport.Vec3{X: 0.8, Y: 0.6, Z: 0.6},
						Fuzz:   0.0,
					},
				},
			},
		},
		ImageTextures: map[string]*pb_transport.ImageTextureMetadata{
			"rusty-metal_albedo.png": {
				Filename: "rusty-metal_albedo.png",
			},
			"rusty-metal_roughness.png": {
				Filename: "rusty-metal_roughness.png",
			},
			"rusty-metal_metallic.png": {
				Filename: "rusty-metal_metallic.png",
			},
			"rusty-metal_normal-ogl.png": {
				Filename: "rusty-metal_normal-ogl.png",
			},
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: sceneTriangles,
			Spheres: []*pb_transport.Sphere{
				{
					Center:       &pb_transport.Vec3{X: -0, Y: 40, Z: 40},
					Radius:       15,
					MaterialName: "light",
				},
			},
		},
	}

	/*
		white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
		light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 15, Y: 15, Z: 15}))
		carMat := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.8, Y: 0.4, Z: 0.4}))
		metal := material.NewMetal(&vec3.Vec3Impl{X: 0.8, Y: .6, Z: .6}, 0)
		//	glass := material.NewDielectric(1.5)
		//	glassSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, &vec3.Vec3Impl{X: -9, Y: 0, Z: 3}, 0, 1, 4, glass)
		//	metal := material.NewMetal(&vec3.Vec3Impl{X: 0.5, Y: 1.0, Z: 1.0}, 0)
		//	metalSphere := hitable.NewSphere(&vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, &vec3.Vec3Impl{X: -24, Y: -4, Z: 6}, 0, 1, 3, metal)
		hitables := []hitable.Hitable{
			hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: 50, Y: 0, Z: 50}, 1, 0, 0, 0, 1, 1, white),
			hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: 50}, &vec3.Vec3Impl{X: 50, Y: 0, Z: 50}, 0, 0, 0, 1, 1, 1, white),
			hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: 50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: -50, Y: 0, Z: -50}, &vec3.Vec3Impl{X: 50, Y: 50, Z: 50}, 1, 0, 0, 0, 1, 1, metal),
			hitable.NewTriangleWithUV(&vec3.Vec3Impl{X: -50, Y: -5, Z: -50}, &vec3.Vec3Impl{X: -50, Y: -5, Z: 50}, &vec3.Vec3Impl{X: -50, Y: 85, Z: -50}, 1, 0, 0, 0, 1, 1, metal),
			hitable.NewSphere(&vec3.Vec3Impl{X: -0, Y: 40, Z: 40}, &vec3.Vec3Impl{X: 0, Y: 40, Z: 40}, 0, 1, 15, light),
		}

		for i := range challengerObj.Groups {
			challengerHitables, err := challengerObj.GroupToHitablesWithCustomMaterial(i, carMat)
			if err != nil {
				return nil, err
			}
			bvh := hitable.NewBVH(challengerHitables, 0, 1)
			hitables = append(hitables, bvh)
		}

		lights := []hitable.Hitable{}
		for _, h := range hitables {
			if h.IsEmitter() {
				lights = append(lights, h)
			}
		}

		lookFrom := &vec3.Vec3Impl{X: 18.0, Y: 8, Z: 12.0}
		lookAt := &vec3.Vec3Impl{X: -6, Y: .7, Z: 0}
		//lookFrom := &vec3.Vec3Impl{X: 2.0, Y: 0.6, Z: 2.0}
		//lookAt := &vec3.Vec3Impl{X: .4, Y: 0, Z: 1}
		vup := &vec3.Vec3Impl{Y: 1}
		distToFocus := 10.0
		aperture := 0.1
		vfov := float32(75.0)
		time0 := 0.0
		time1 := 1.0
		cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)
	*/

	marshalledScene, err := proto.Marshal(protoScene)
	if err != nil {
		return nil, err
	}
	return marshalledScene, nil

}

func CornellBoxPB(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:    "Cornell Box",
		Version: "1.0.0",
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 33,
					},
					Vertex2: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					Vertex2: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
			},
			Spheres: []*pb_transport.Sphere{
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "Glass",
				},
				{
					Center: &pb_transport.Vec3{
						X: 70,
						Y: 30,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "Rusty Metal",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.73,
											Z: 0.73,
										},
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.0,
											Y: 0.73,
											Z: 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.0,
											Z: 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_Emit{
							Emit: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 15,
											Y: 15,
											Z: 15,
										},
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_Refidx{
							Refidx: 1.5,
						},
					},
				},
			},
			"Marine Blue": {
				Name: "Marine Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{X: 0.0, Y: 0.26666666666666666, Z: 0.5058823529411764},
									},
								},
							},
						},
					},
				},
			},
			"Rusty Metal": {
				Name: "Rusty Metal",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_albedo.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_normal-ogl.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_roughness.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
		},
		ImageTextures: map[string]*pb_transport.ImageTextureMetadata{
			"rusty-metal_albedo.png": {
				Filename: "rusty-metal_albedo.png",
			},
			"rusty-metal_metallic.png": {
				Filename: "rusty-metal_metallic.png",
			},
			"rusty-metal_normal-ogl.png": {
				Filename: "rusty-metal_normal-ogl.png",
			},
			"rusty-metal_roughness.png": {
				Filename: "rusty-metal_roughness.png",
			},
		},
	}

	return protoScene
}

func CornellBoxRGB(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:    "Cornell Box",
		Version: "1.0.0",
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 33,
					},
					Vertex2: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					Vertex2: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
			},
			Spheres: []*pb_transport.Sphere{
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "Glass",
				},
				{
					Center: &pb_transport.Vec3{
						X: 70,
						Y: 30,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "Marine Blue",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.73,
											Z: 0.73,
										},
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.0,
											Y: 0.73,
											Z: 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.0,
											Z: 0.0,
										},
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_Emit{
							Emit: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 15,
											Y: 15,
											Z: 15,
										},
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_Refidx{
							Refidx: 1.5,
						},
					},
				},
			},
			"Marine Blue": {
				Name: "Marine Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{X: 0.0, Y: 0.26666666666666666, Z: 0.5058823529411764},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return protoScene
}

func CornellBoxSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					MaterialName: "White",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 0,
						Y: 100,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 0,
						Y: 0,
						Z: 0,
					},
					MaterialName: "Green",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 0,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 100,
					},
					Vertex1: &pb_transport.Vec3{
						X: 100,
						Y: 100,
						Z: 100,
					},
					Vertex2: &pb_transport.Vec3{
						X: 100,
						Y: 0,
						Z: 0,
					},
					Uv0: &pb_transport.Vec2{
						U: 0,
						V: 0,
					},
					Uv1: &pb_transport.Vec2{
						U: 1,
						V: 0,
					},
					Uv2: &pb_transport.Vec2{
						U: 1,
						V: 1,
					},
					MaterialName: "Red",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 33,
					},
					Vertex2: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
				{
					Vertex0: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 33,
					},
					Vertex1: &pb_transport.Vec3{
						X: 66,
						Y: 99,
						Z: 66,
					},
					Vertex2: &pb_transport.Vec3{
						X: 33,
						Y: 99,
						Z: 66,
					},
					MaterialName: "white_light",
				},
			},
			Spheres: []*pb_transport.Sphere{
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "Glass",
				},
				{
					Center: &pb_transport.Vec3{
						X: 70,
						Y: 30,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "Rusty Metal",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 540.0, // Slightly shifted for more saturated green
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 640.0, // Slightly shifted for more saturated red
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 15.0, // High emission for light
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for typical glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
					},
				},
			},
			"Marine Blue": {
				Name: "Marine Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.5,
										CenterWavelength: 480.0, // Blue wavelength
										Width:            80.0,  // Broader blue response
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

func CornellBoxPrismSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Prism Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Room walls (same as CornellBoxSpectral)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Additional floor triangle
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Red",
				},
				// Light source (same as CornellBoxSpectral)
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				// Triangular Prism (replacing the glass sphere) - Bigger, lifted, and rotated
				// Base triangle (bottom face) - rotated and tilted
				{
					Vertex0:      &pb_transport.Vec3{X: 40, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 50, Y: 30, Z: 10}, // Bottom right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 30, Z: 40}, // Top
					MaterialName: "Glass",
				},
				// Top triangle (top face) - rotated and tilted
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 50, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 50, Z: 35}, // Bottom right
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 50, Z: 50}, // Top
					MaterialName: "Glass",
				},
				// Side faces (3 rectangles, each made of 2 triangles)
				// Left face
				{
					Vertex0:      &pb_transport.Vec3{X: 40, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 15, Y: 50, Z: 35}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 50, Z: 50}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 40, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 30, Y: 50, Z: 50}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 30, Z: 40}, // Bottom right
					MaterialName: "Glass",
				},
				// Right face
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 50, Z: 35}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 50, Z: 50}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 30, Y: 50, Z: 50}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 30, Z: 40}, // Bottom right
					MaterialName: "Glass",
				},
				// Back face
				{
					Vertex0:      &pb_transport.Vec3{X: 40, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 15, Y: 50, Z: 35}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 45, Y: 50, Z: 35}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 40, Y: 30, Z: 10}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 50, Z: 35}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 50, Y: 30, Z: 10}, // Bottom right
					MaterialName: "Glass",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Keep the Marine Blue sphere
				{
					Center: &pb_transport.Vec3{
						X: 70,
						Y: 30,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "Marine Blue",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 540.0, // Slightly shifted for more saturated green
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 640.0, // Slightly shifted for more saturated red
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 15.0, // High emission for light
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for typical glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
					},
				},
			},
			"Marine Blue": {
				Name: "Marine Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.5,
										CenterWavelength: 480.0, // Blue wavelength
										Width:            80.0,  // Broader blue response
									},
								},
							},
						},
					},
				},
			},
			"Rusty Metal": {
				Name: "Rusty Metal",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_albedo.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_normal-ogl.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "rusty-metal_roughness.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
		},
		ImageTextures: map[string]*pb_transport.ImageTextureMetadata{
			"rusty-metal_albedo.png": {
				Filename: "rusty-metal_albedo.png",
			},
			"rusty-metal_metallic.png": {
				Filename: "rusty-metal_metallic.png",
			},
			"rusty-metal_normal-ogl.png": {
				Filename: "rusty-metal_normal-ogl.png",
			},
			"rusty-metal_roughness.png": {
				Filename: "rusty-metal_roughness.png",
			},
		},
	}

	return protoScene
}

func CornellBoxPrismSpectralEnhanced(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Prism Spectral Enhanced",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Additional floor triangle
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Red",
				},
				// Light source (same as CornellBoxSpectral)
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				// Triangular Prism (replacing the glass sphere) - Bigger, lifted, and rotated around Z-axis (horizontal orientation)
				// Base triangle (bottom face) - rotated around Z-axis for horizontal orientation
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 55, Y: 15, Z: 35}, // Bottom right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 15, Z: 50}, // Top
					MaterialName: "Glass",
				},
				// Top triangle (top face) - rotated around Z-axis for horizontal orientation
				{
					Vertex0:      &pb_transport.Vec3{X: 10, Y: 55, Z: 30}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 50, Y: 55, Z: 30}, // Bottom right
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 55, Z: 60}, // Top
					MaterialName: "Glass",
				},
				// Side faces (3 rectangles, each made of 2 triangles)
				// Left face
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 10, Y: 55, Z: 30}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 55, Z: 60}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 30, Y: 55, Z: 60}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 15, Z: 50}, // Bottom right
					MaterialName: "Glass",
				},
				// Right face
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 50, Y: 55, Z: 30}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 30, Y: 55, Z: 60}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 30, Y: 55, Z: 60}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 35, Y: 15, Z: 50}, // Bottom right
					MaterialName: "Glass",
				},
				// Back face
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 10, Y: 55, Z: 30}, // Top left
					Vertex2:      &pb_transport.Vec3{X: 50, Y: 55, Z: 30}, // Top right
					MaterialName: "Glass",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 15, Y: 15, Z: 35}, // Bottom left
					Vertex1:      &pb_transport.Vec3{X: 50, Y: 55, Z: 30}, // Top right
					Vertex2:      &pb_transport.Vec3{X: 55, Y: 15, Z: 35}, // Bottom right
					MaterialName: "Glass",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Keep the Marine Blue sphere
				{
					Center: &pb_transport.Vec3{
						X: 70,
						Y: 30,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "Marine Blue",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 540.0, // Slightly shifted for more saturated green
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Higher peak for more vibrant color
										CenterWavelength: 640.0, // Slightly shifted for more saturated red
										Width:            40.0,  // Narrower for more saturation
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 15.0, // High emission for light
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for high-dispersion glass (dramatic dispersion for visible chromatic aberration)
										Values: []float32{1.65, 1.62, 1.60, 1.58, 1.56, 1.54, 1.52, 1.50, 1.48, 1.46, 1.44, 1.42, 1.40, 1.38, 1.36, 1.34, 1.32, 1.30, 1.28, 1.26},
									},
								},
							},
						},
					},
				},
			},
			"Marine Blue": {
				Name: "Marine Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.5,
										CenterWavelength: 480.0, // Blue wavelength
										Width:            80.0,  // Broader blue response
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

func CornellBoxDiamondsSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Diamonds Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 60,
				Z: -120,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 30,
				Z: 50,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      35,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Additional floor triangle
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Red",
				},
				// Light source (larger and brighter for diamonds)
				{
					Vertex0:      &pb_transport.Vec3{X: 25, Y: 99, Z: 25},
					Vertex1:      &pb_transport.Vec3{X: 75, Y: 99, Z: 25},
					Vertex2:      &pb_transport.Vec3{X: 75, Y: 99, Z: 75},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 25, Y: 99, Z: 25},
					Vertex1:      &pb_transport.Vec3{X: 75, Y: 99, Z: 75},
					Vertex2:      &pb_transport.Vec3{X: 25, Y: 99, Z: 75},
					MaterialName: "white_light",
				},
				// Diamond 1 - Large main diamond (enhanced brilliant cut with more facets) - Rotated 15° around X, 15° around Y, 15° around Z
				// Crown (top) facets - 8 main crown facets
				{
					Vertex0:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 55, Y: 22, Z: 38}, // Right top
					Vertex2:      &pb_transport.Vec3{X: 48, Y: 32, Z: 52}, // Front top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 48, Y: 32, Z: 52}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 48, Y: 32, Z: 52}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 62, Y: 18, Z: 62}, // Back top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 62, Y: 18, Z: 62}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 48, Y: 32, Z: 52}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 55, Y: 22, Z: 38}, // Right top
					MaterialName: "Diamond",
				},
				// Additional crown facets for more sparkle
				{
					Vertex0:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 52, Y: 23, Z: 47}, // Front right
					Vertex2:      &pb_transport.Vec3{X: 43, Y: 27, Z: 52}, // Front left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 43, Y: 27, Z: 52}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 43, Y: 27, Z: 52}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 53, Y: 23, Z: 57}, // Back left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 62, Y: 18, Z: 62}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 53, Y: 23, Z: 57}, // Back left
					Vertex2:      &pb_transport.Vec3{X: 57, Y: 33, Z: 52}, // Back right
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 22, Z: 38}, // Right top
					Vertex1:      &pb_transport.Vec3{X: 57, Y: 33, Z: 52}, // Back right
					Vertex2:      &pb_transport.Vec3{X: 52, Y: 23, Z: 47}, // Front right
					MaterialName: "Diamond",
				},
				// Pavilion (bottom) facets - 8 main pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 55, Y: 22, Z: 38}, // Right bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 55, Y: 22, Z: 38}, // Right bottom
					Vertex2:      &pb_transport.Vec3{X: 62, Y: 18, Z: 62}, // Back bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 62, Y: 18, Z: 62}, // Back bottom
					Vertex2:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 38, Y: 22, Z: 55}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 45, Y: 18, Z: 45}, // Front bottom
					MaterialName: "Diamond",
				},
				// Additional pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 48, Y: 17, Z: 47}, // Front center
					Vertex2:      &pb_transport.Vec3{X: 43, Y: 17, Z: 52}, // Left center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 43, Y: 17, Z: 52}, // Left center
					Vertex2:      &pb_transport.Vec3{X: 52, Y: 17, Z: 57}, // Back center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 52, Y: 17, Z: 57}, // Back center
					Vertex2:      &pb_transport.Vec3{X: 57, Y: 17, Z: 52}, // Right center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 50, Y: 6, Z: 50},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 57, Y: 17, Z: 52}, // Right center
					Vertex2:      &pb_transport.Vec3{X: 48, Y: 17, Z: 47}, // Front center
					MaterialName: "Diamond",
				},
				// Diamond 2 - Medium diamond (left) - enhanced with more facets - Rotated 15° around X, -15° around Y, 0° around Z
				// Crown facets - 8 main crown facets
				{
					Vertex0:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 17, Z: 28}, // Right top
					Vertex2:      &pb_transport.Vec3{X: 28, Y: 24, Z: 37}, // Front top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 28, Y: 24, Z: 37}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 28, Y: 24, Z: 37}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 38, Y: 17, Z: 45}, // Back top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 38, Y: 17, Z: 45}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 28, Y: 24, Z: 37}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 45, Y: 17, Z: 28}, // Right top
					MaterialName: "Diamond",
				},
				// Additional crown facets
				{
					Vertex0:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 32, Y: 16, Z: 30}, // Front right
					Vertex2:      &pb_transport.Vec3{X: 23, Y: 30, Z: 37}, // Front left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 23, Y: 30, Z: 37}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 23, Y: 30, Z: 37}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 16, Z: 40}, // Back left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 38, Y: 17, Z: 45}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 33, Y: 16, Z: 40}, // Back left
					Vertex2:      &pb_transport.Vec3{X: 37, Y: 30, Z: 37}, // Back right
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 45, Y: 17, Z: 28}, // Right top
					Vertex1:      &pb_transport.Vec3{X: 37, Y: 30, Z: 37}, // Back right
					Vertex2:      &pb_transport.Vec3{X: 32, Y: 16, Z: 30}, // Front right
					MaterialName: "Diamond",
				},
				// Pavilion facets - 8 main pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 45, Y: 17, Z: 28}, // Right bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 45, Y: 17, Z: 28}, // Right bottom
					Vertex2:      &pb_transport.Vec3{X: 38, Y: 17, Z: 45}, // Back bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 38, Y: 17, Z: 45}, // Back bottom
					Vertex2:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 22, Y: 13, Z: 38}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 18, Y: 13, Z: 32}, // Front bottom
					MaterialName: "Diamond",
				},
				// Additional pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 28, Y: 14, Z: 30}, // Front center
					Vertex2:      &pb_transport.Vec3{X: 23, Y: 14, Z: 37}, // Left center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 23, Y: 14, Z: 37}, // Left center
					Vertex2:      &pb_transport.Vec3{X: 32, Y: 14, Z: 40}, // Back center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 32, Y: 14, Z: 40}, // Back center
					Vertex2:      &pb_transport.Vec3{X: 37, Y: 14, Z: 37}, // Right center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 30, Y: 4, Z: 35},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 37, Y: 14, Z: 37}, // Right center
					Vertex2:      &pb_transport.Vec3{X: 28, Y: 14, Z: 30}, // Front center
					MaterialName: "Diamond",
				},
				// Diamond 3 - Small diamond (right) - enhanced with more facets - Rotated 0° around X, 15° around Y, 15° around Z
				// Crown facets - 8 main crown facets
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 82, Y: 10, Z: 53}, // Right top
					Vertex2:      &pb_transport.Vec3{X: 68, Y: 16, Z: 62}, // Front top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 68, Y: 16, Z: 62}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 68, Y: 16, Z: 62}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 78, Y: 10, Z: 67}, // Back top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 78, Y: 10, Z: 67}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 68, Y: 16, Z: 62}, // Front top
					Vertex2:      &pb_transport.Vec3{X: 82, Y: 10, Z: 53}, // Right top
					MaterialName: "Diamond",
				},
				// Additional crown facets
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 72, Y: 13, Z: 59}, // Front right
					Vertex2:      &pb_transport.Vec3{X: 63, Y: 13, Z: 62}, // Front left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Center top
					Vertex1:      &pb_transport.Vec3{X: 63, Y: 13, Z: 62}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left top
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left top
					Vertex1:      &pb_transport.Vec3{X: 63, Y: 13, Z: 62}, // Front left
					Vertex2:      &pb_transport.Vec3{X: 72, Y: 13, Z: 65}, // Back left
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 78, Y: 10, Z: 67}, // Back top
					Vertex1:      &pb_transport.Vec3{X: 72, Y: 13, Z: 65}, // Back left
					Vertex2:      &pb_transport.Vec3{X: 77, Y: 13, Z: 62}, // Back right
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 82, Y: 10, Z: 53}, // Right top
					Vertex1:      &pb_transport.Vec3{X: 77, Y: 13, Z: 62}, // Back right
					Vertex2:      &pb_transport.Vec3{X: 72, Y: 13, Z: 59}, // Front right
					MaterialName: "Diamond",
				},
				// Pavilion facets - 8 main pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 82, Y: 10, Z: 53}, // Right bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 82, Y: 10, Z: 53}, // Right bottom
					Vertex2:      &pb_transport.Vec3{X: 78, Y: 10, Z: 67}, // Back bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 78, Y: 10, Z: 67}, // Back bottom
					Vertex2:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left bottom
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60},  // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 62, Y: 10, Z: 63}, // Left bottom
					Vertex2:      &pb_transport.Vec3{X: 55, Y: 10, Z: 57}, // Front bottom
					MaterialName: "Diamond",
				},
				// Additional pavilion facets
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60}, // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 68, Y: 8, Z: 59}, // Front center
					Vertex2:      &pb_transport.Vec3{X: 63, Y: 8, Z: 62}, // Left center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60}, // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 63, Y: 8, Z: 62}, // Left center
					Vertex2:      &pb_transport.Vec3{X: 72, Y: 8, Z: 65}, // Back center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60}, // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 72, Y: 8, Z: 65}, // Back center
					Vertex2:      &pb_transport.Vec3{X: 77, Y: 8, Z: 62}, // Right center
					MaterialName: "Diamond",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 70, Y: 2, Z: 60}, // Center bottom
					Vertex1:      &pb_transport.Vec3{X: 77, Y: 8, Z: 62}, // Right center
					Vertex2:      &pb_transport.Vec3{X: 68, Y: 8, Z: 59}, // Front center
					MaterialName: "Diamond",
				},
			},
			Spheres: []*pb_transport.Sphere{},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 540.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 640.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 25.0, // Brighter light for diamonds
									},
								},
							},
						},
					},
				},
			},
			"Diamond": {
				Name: "Diamond",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for diamond (extremely high dispersion)
										Values: []float32{2.45, 2.42, 2.40, 2.38, 2.36, 2.34, 2.32, 2.30, 2.28, 2.26, 2.24, 2.22, 2.20, 2.18, 2.16, 2.14, 2.12, 2.10, 2.08, 2.06},
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

// CornellBoxColoredGlassSpectral creates a Cornell Box scene with colored glass spheres
// demonstrating the Beer-Lambert law for spectral absorption
func CornellBoxColoredGlassSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Colored Glass Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -120,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 50,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      35,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Light (white_light)
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "Red",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Pyramid of colored glass spheres (rotated ~30 degrees around Y-axis)
				// Bottom layer: 6 red spheres arranged in a triangle pattern
				// Row 1: 3 spheres
				{
					Center:       &pb_transport.Vec3{X: 30, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "RedGlass",
				},
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "RedGlass",
				},
				{
					Center:       &pb_transport.Vec3{X: 70, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "RedGlass",
				},
				// Row 2: 2 spheres
				{
					Center:       &pb_transport.Vec3{X: 40, Y: 15, Z: 50},
					Radius:       10,
					MaterialName: "RedGlass",
				},
				{
					Center:       &pb_transport.Vec3{X: 60, Y: 15, Z: 50},
					Radius:       10,
					MaterialName: "RedGlass",
				},
				// Row 3: 1 sphere
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 15, Z: 70},
					Radius:       10,
					MaterialName: "RedGlass",
				},

				// Middle layer: 3 green spheres positioned in gaps between red spheres
				{
					Center:       &pb_transport.Vec3{X: 40, Y: 28, Z: 40},
					Radius:       10,
					MaterialName: "GreenGlass",
				},
				{
					Center:       &pb_transport.Vec3{X: 60, Y: 28, Z: 40},
					Radius:       10,
					MaterialName: "GreenGlass",
				},
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 28, Z: 60},
					Radius:       10,
					MaterialName: "GreenGlass",
				},

				// Top layer: 1 blue sphere
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 42, Z: 50},
					Radius:       10,
					MaterialName: "BlueGlass",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 540.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 640.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 15.0,
									},
								},
							},
						},
					},
				},
			},
			"BlueGlass": {
				Name: "BlueGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.1,   // Moderate absorption at blue wavelengths
										CenterWavelength: 480.0, // Blue wavelength
										Width:            60.0,  // Broad absorption in blue region
									},
								},
							},
						},
					},
				},
			},
			"RedGlass": {
				Name: "RedGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.1,   // Moderate absorption at red wavelengths
										CenterWavelength: 640.0, // Red wavelength
										Width:            60.0,  // Broad absorption in red region
									},
								},
							},
						},
					},
				},
			},
			"GreenGlass": {
				Name: "GreenGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.1,   // Moderate absorption at green wavelengths
										CenterWavelength: 540.0, // Green wavelength
										Width:            60.0,  // Broad absorption in green region
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

// CornellBoxWaterSpectral creates a Cornell Box scene with water filling 40% of the box
// and a colored glass sphere floating in it to demonstrate refraction
func CornellBoxWaterSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Water Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -120,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 50,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      35,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Light (white_light)
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "Red",
				},
				// Water cube (40% of box height = 40 units, extending to walls)
				// IMPORTANT: The water body must extend exactly to the Cornell Box walls (X=0-100, Z=0-100)
				// with only a small Y offset (Y=0.01) to avoid floor intersection. This is critical because:
				// 1. If there are gaps between water and walls, the walls won't be visible through the water
				// 2. If the water intersects with walls/floor, it causes numerical precision issues and rendering artifacts
				// 3. The small Y offset (0.01) prevents floor intersection while keeping gaps negligible
				// 4. This construction ensures proper light propagation through the water body
				// Water cube bottom face
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 100},
					MaterialName: "Water",
				},
				// Water cube top face (at 40% height = 40 units)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				// Water cube front face
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				// Water cube back face
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				// Water cube left face
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				// Water cube right face
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0.01, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					MaterialName: "Water",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Lambertian sphere floating above water
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 40, Z: 50},
					Radius:       15,
					MaterialName: "BlueLambert",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 540.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 640.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 15.0,
									},
								},
							},
						},
					},
				},
			},
			"Water": {
				Name: "Water",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for water (slight dispersion)
										Values: []float32{1.344, 1.343, 1.342, 1.341, 1.340, 1.339, 1.338, 1.337, 1.336, 1.335, 1.334, 1.333, 1.332, 1.331, 1.330, 1.329, 1.328, 1.327, 1.326, 1.325},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.0, // Clean water has minimal absorption in visible spectrum
									},
								},
							},
						},
					},
				},
			},
			"BlueGlass": {
				Name: "BlueGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.1,   // Moderate absorption at blue wavelengths
										CenterWavelength: 480.0, // Blue wavelength
										Width:            60.0,  // Broad absorption in blue region
									},
								},
							},
						},
					},
				},
			},
			"BlueLambert": {
				Name: "BlueLambert",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.8,   // High reflectance at blue wavelengths
										CenterWavelength: 480.0, // Blue wavelength
										Width:            60.0,  // Broad reflectance in blue region
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

// CornellBoxTransparentPyramidSpectral creates a Cornell Box scene with a floating pyramid
// of transparent spheres to demonstrate refraction and transparency effects
func CornellBoxTransparentPyramidSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Transparent Pyramid Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -120,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 50,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      35,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Ceiling (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Light (white_light)
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "Red",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Pyramid of transparent spheres on the floor
				// Bottom layer: 6 transparent spheres arranged in a triangle pattern
				// Row 1: 3 spheres
				{
					Center:       &pb_transport.Vec3{X: 30, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "Transparent",
				},
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "Transparent",
				},
				{
					Center:       &pb_transport.Vec3{X: 70, Y: 15, Z: 30},
					Radius:       10,
					MaterialName: "Transparent",
				},
				// Row 2: 2 spheres
				{
					Center:       &pb_transport.Vec3{X: 40, Y: 15, Z: 50},
					Radius:       10,
					MaterialName: "Transparent",
				},
				{
					Center:       &pb_transport.Vec3{X: 60, Y: 15, Z: 50},
					Radius:       10,
					MaterialName: "Transparent",
				},
				// Row 3: 1 sphere
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 15, Z: 70},
					Radius:       10,
					MaterialName: "Transparent",
				},

				// Middle layer: 3 transparent spheres positioned in gaps between bottom spheres
				{
					Center:       &pb_transport.Vec3{X: 40, Y: 28, Z: 40},
					Radius:       10,
					MaterialName: "Transparent",
				},
				{
					Center:       &pb_transport.Vec3{X: 60, Y: 28, Z: 40},
					Radius:       10,
					MaterialName: "Transparent",
				},
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 28, Z: 60},
					Radius:       10,
					MaterialName: "Transparent",
				},

				// Top layer: 1 transparent sphere
				{
					Center:       &pb_transport.Vec3{X: 50, Y: 42, Z: 50},
					Radius:       10,
					MaterialName: "Transparent",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 540.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,
										CenterWavelength: 640.0,
										Width:            40.0,
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 1.0,
									},
								},
							},
						},
					},
				},
			},
			// Transparent material with minimal absorption for clear glass effect
			"Transparent": {
				Name: "Transparent",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.01, // Very low absorption for transparency
									},
								},
							},
						},
					},
				},
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}
