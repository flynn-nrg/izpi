// Package colours defines a few useful colours.
package colours

import "github.com/flynn-nrg/izpi/pkg/vec3"

var (
	Red            = &vec3.Vec3Impl{X: 1.0}
	Green          = &vec3.Vec3Impl{Y: 1.0}
	Blue           = &vec3.Vec3Impl{Z: 1.0}
	Black          = &vec3.Vec3Impl{}
	White          = &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}
	RoyalBlue      = &vec3.Vec3Impl{X: 48.0 / 255.0, Y: 87.0 / 255.0, Z: 225.0 / 255.0}
	ResolutionBlue = &vec3.Vec3Impl{X: 0.0, Y: 32.0 / 255.0, Z: 130.0 / 255.0}
	LimeStone      = &vec3.Vec3Impl{X: 152.0 / 255.0, Y: 154.0 / 255.0, Z: 152.0 / 255.0}
)
