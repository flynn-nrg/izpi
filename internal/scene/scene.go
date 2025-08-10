// Package scene implements structures and methods to work with scenes.
package scene

import (
	"errors"

	"github.com/flynn-nrg/izpi/internal/camera"
	"github.com/flynn-nrg/izpi/internal/hitable"
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
