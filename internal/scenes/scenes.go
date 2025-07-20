// Package scenes implements some sample scenes.
package scenes

import (
	"fmt"
	"math/rand"
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
			chooseMat := rand.Float64()
			center := &vec3.Vec3Impl{X: float64(a) + 0.9*rand.Float64(), Y: 0.2, Z: float64(b) + 0.9*rand.Float64()}
			if vec3.Sub(center, &vec3.Vec3Impl{X: 4, Y: 0.2, Z: 0}).Length() > 0.9 {
				if chooseMat < 0.8 {
					// diffuse
					spheres = append(spheres, hitable.NewSphere(center,
						vec3.Add(center, &vec3.Vec3Impl{Y: 0.5 * rand.Float64()}), 0.0, 1.0, 0.2,
						material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{
							X: rand.Float64() * rand.Float64(),
							Y: rand.Float64() * rand.Float64(),
							Z: rand.Float64() * rand.Float64(),
						}))))
				} else if chooseMat < 0.95 {
					// metal
					spheres = append(spheres, hitable.NewSphere(center, center, 0.0, 1.0, 0.2,
						material.NewMetal(&vec3.Vec3Impl{
							X: 0.5 * (1.0 - rand.Float64()),
							Y: 0.5 * (1.0 - rand.Float64()),
							Z: 0.5 * (1.0 - rand.Float64()),
						}, 0.2*rand.Float64())))
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
func CornellBox(aspect float64) *scene.Scene {
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
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam)
}

// Final returns the scene from the last chapter in the book.
func Final(aspect float64) (*hitable.HitableSlice, *camera.Camera) {
	nb := 20
	list := []hitable.Hitable{}
	boxList := []hitable.Hitable{}
	boxList2 := []hitable.Hitable{}

	white := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.73, Y: 0.73, Z: 0.73}))
	ground := material.NewLambertian(texture.NewConstant(&vec3.Vec3Impl{X: 0.48, Y: 0.83, Z: 0.53}))

	for i := 0; i < nb; i++ {
		for j := 0; j < nb; j++ {
			w := float64(100)
			x0 := -1000.0 + float64(i)*w
			z0 := -1000.0 + float64(j)*w
			y0 := float64(0)
			x1 := x0 + w
			y1 := 100.0 * (rand.Float64() + 0.01)
			z1 := z0 + w
			boxList = append(boxList, hitable.NewBox(&vec3.Vec3Impl{X: x0, Y: y0, Z: z0}, &vec3.Vec3Impl{X: x1, Y: y1, Z: z1}, ground))
		}
	}

	list = append(list, hitable.NewBVH(boxList, 0, 1))

	light := material.NewDiffuseLight(texture.NewConstant(&vec3.Vec3Impl{X: 7, Y: 7, Z: 7}))
	list = append(list, hitable.NewXZRect(123, 423, 147, 412, 554, light))

	center := &vec3.Vec3Impl{X: 400, Y: 400, Z: 200}
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
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 400, Y: 200, Z: 400}, &vec3.Vec3Impl{X: 400, Y: 200, Z: 400}, 0, 1, 100, emat))

	perText := texture.NewNoise(0.1)
	list = append(list, hitable.NewSphere(&vec3.Vec3Impl{X: 220, Y: 280, Z: 300}, &vec3.Vec3Impl{X: 220, Y: 280, Z: 300}, 0, 1, 80, material.NewLambertian(perText)))

	ns := 1000
	for j := 0; j < ns; j++ {
		center := &vec3.Vec3Impl{X: 165 * rand.Float64(), Y: 165 * rand.Float64(), Z: 165 * rand.Float64()}
		boxList2 = append(boxList2, hitable.NewSphere(center, center, 0, 1, 10, white))
	}

	list = append(list, hitable.NewTranslate(hitable.NewRotateY(hitable.NewBVH(boxList2, 0, 1), 15), &vec3.Vec3Impl{X: -100, Y: 270, Z: 395}))

	lookFrom := &vec3.Vec3Impl{X: 478.0, Y: 278.0, Z: -600.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return hitable.NewSlice(list), cam
}

// Environment returns a scene that tests the sky dome and HDR textures functionality.
func Environment(aspect float64) *scene.Scene {
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
	vfov := float64(60.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam)

}

// CornellBox returns a scene recreating the Cornell box.
func CornellBoxObj(aspect float64) (*scene.Scene, error) {
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

	cube.Translate(&vec3.Vec3Impl{X: 280, Y: 20, Z: 390})
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
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

func TVSet(aspect float64) (*scene.Scene, error) {
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
	cube.Translate(&vec3.Vec3Impl{X: 280, Y: 100, Z: 420})

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
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

func SWHangar(aspect float64) (*scene.Scene, error) {
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
		hitable.NewSphere(&vec3.Vec3Impl{X: 0, Y: 20, Z: -30}, &vec3.Vec3Impl{X: 0, Y: 20, Z: -30}, 0, 1, 20, light),
		hitable.NewSphere(&vec3.Vec3Impl{X: -50, Y: 15, Z: 80}, &vec3.Vec3Impl{X: -50, Y: 15, Z: 80}, 0, 1, 20, glass),
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

	lookFrom := &vec3.Vec3Impl{X: 0.0, Y: 25.0, Z: 30.0}
	lookAt := &vec3.Vec3Impl{X: -1, Y: 25, Z: 31}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	vfov := float64(100.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

// DisplacementTest returns a scene recreating the Cornell box and a displacement map applied to the floor.
func DisplacementTest(aspect float64) (*scene.Scene, error) {
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
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil
}

// VWBeetle returns a scene that tests the sky dome and HDR textures functionality.
func VWBeetle(aspect float64) (*scene.Scene, error) {
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
	vfov := float64(60.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	return scene.New(hitable.NewSlice(hitables), hitable.NewSlice(lights), cam), nil

}

// Challenger returns a scene that tests the sky dome and HDR textures functionality.
func Challenger(aspect float64) ([]byte, error) {
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
		vfov := float64(75.0)
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

func CornellBoxPB(aspect float64) *pb_transport.Scene {
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
						Y: 20,
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

func CornellBoxRGB(aspect float64) *pb_transport.Scene {
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
						Y: 20,
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

func CornellBoxSpectral(aspect float64) *pb_transport.Scene {
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
						Y: 20,
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
										PeakValue:        0.73,
										CenterWavelength: 550.0, // Green wavelength
										Width:            50.0,  // Narrow green response
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
										PeakValue:        0.73,
										CenterWavelength: 650.0, // Red wavelength
										Width:            50.0,  // Narrow red response
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
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 1.5, // Glass refractive index
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
