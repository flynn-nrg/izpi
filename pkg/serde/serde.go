// Package serde implements functions to serialise and deserialise scene data.
package serde

import (
	"io"
)

// Serde defines the methods to seralise and deserialise scene data.
type Serde interface {
	// Serialise writes a serialised version of the data to the provided writer.
	Serialise(scene *Scene, w io.Writer) error
	// Deserialise returns a struct representation of the serialised data.
	Deserialise(r io.Reader) (*Scene, error)
}

// Vec3 represents a vector.
type Vec3 struct {
	// X is the x coordinate of this vector.
	X float64
	// Y is the y coordinate of this vector.
	Y float64
	// Z is the z coordinate of this vector.
	Z float64
}

// Camera represents a camera.
type Camera struct {
	// LookFrom is the location of the camera.
	LookFrom Vec3
	// LookAt is where the camera is pointing at.
	LookAt Vec3
	// VUp defines the "up" vector.
	VUp Vec3
	// VFov define the field of view.
	VFov float64
	// Aspect is the aspect ratio.
	Aspect float64 `yaml:"Aspect,omitempty"`
	// Aperture is the lens aperture for this camera.
	Aperture float64
	// FocusDist is the focus distance.
	FocusDist float64
	// Time0 defines the beginning time of the exposure.
	Time0 float64
	// Time1 defines the end time of the exposure.
	Time1 float64
}

// Image represents an image based texture.
type Image struct {
	// FileName is the name of the file containing the texture data.
	FileName string
}

// Constant represents a constant texture.
type Constant struct {
	// Value is a vector with the constant data.
	Value Vec3
}

// Noise represents a Perlin noise texture.
type Noise struct {
	// Scale defines the Perlin noise scale.
	Scale float64
}

// Texture represents a texture instance.
type Texture struct {
	// Name is the name of the texture.
	Name string
	// Type is the texture type: constant, image or noise.
	Type string
	// Image is an instance of an Image texture.
	Image Image `yaml:"Image,omitempty"`
	// Constant is an instance of a Constant texture.
	Constant Constant `yaml:"Constant,omitempty"`
	// Noise is an instance of a Perlin noise texture.
	Noise Noise `yaml:"Noise,omitempty"`
}

// Lambert represents a Lambertian material.
type Lambert struct {
	// Albedo is the colour texture.
	Albedo Texture
}

// DiffuseLight represents a diffuse light material.
type DiffuseLight struct {
	// Emit is the colour texture.
	Emit Texture
}

// Isotropic represents a dielectric material.
type Isotropic struct {
	// Albedo is the colour texture.
	Albedo Texture
}

// Metal represents a metallic material.
type Metal struct {
	// Albedo is the colour texture.
	Albedo Texture
	// Fuzz defines how shiny a metallic surface is. 0 is a perfect mirror.
	Fuzz float64
}

// Dielectric represents a dielectric material.
type Dielectric struct {
	// RefIdx is the refraction index.
	RefIdx float64
}

// Material represents a single material.
type Material struct {
	// Name is the name of the material.
	Name string
	// Type is the type of material: lambert, diffuse_light, isotropic, metal, dielectric.
	Type string
	// Lambert is a lambert material.
	Lambert Lambert `yaml:"Lambert,omitempty"`
	// DiffuseLight is a diffuse light.
	DiffuseLight DiffuseLight `yaml:"DiffuseLight,omitempty"`
	// Isotropic is an isotropic material.
	Isotropic Isotropic `yaml:"Isotropic,omitempty"`
	// Metal is a metallic material.
	Metal Metal `yaml:"Metal,omitempty"`
	// Dielectric is a dielectric material.
	Dielectric Dielectric `yaml:"Dielectric,omitempty"`
}

// Sphere represents a sphere.
type Sphere struct {
	// Center defines the centre of the sphere.
	Center Vec3
	// Radius is the radius of the sphere.
	Radius float64
	// Material is the material of the sphere.
	Material Material
}

// Triangle represents a single triangle.
// Vertices must be defined counter clockwise.
type Triangle struct {
	// Vertex0 is the first vertex of this triangle.
	Vertex0 Vec3
	// Vertex1 is the second vertex of this triangle.
	Vertex1 Vec3
	// Vertex2 is the third vertex of this triangle.
	Vertex2 Vec3
	// U0 is the u coordinate of the first vertex.
	U0 float64 `yaml:"U0,omitempty"`
	// V0 is the v coordinate of the first vertex.
	V0 float64 `yaml:"V0,omitempty"`
	// U1 is the u coordinate of the second vertex.
	U1 float64 `yaml:"U1,omitempty"`
	// V1 is the v coordinate of the second vertex.
	V1 float64 `yaml:"V1,omitempty"`
	// U2 is the u coordinate of the third vertex.
	U2 float64 `yaml:"U2,omitempty"`
	// V2 is the v coordinate of the third vertex.
	V2 float64 `yaml:"V2,omitempty"`
	// NormalMap is the normnal map texture associated with this triangle.
	NormalMap Image `yaml:"NormalMap,omitempty"`
	// Material is the material for this triangle.
	Material Material
}

// Mesh represents a Wavefront OBJ instance.
type Mesh struct {
	// WavefrontData is the name of the file containing the mesh information.
	WavefrontData string
	// Translate is the translation vector that is applied to all the vertices.
	Translate Vec3
	// Scale is a the scale vector that is applied to all the vetices.
	Scale Vec3
	// NormalMap is the normal map that gets applied to this mesh.
	NormalMap Image `yaml:"NormalMap,omitempty"`
	// Material is the material associated with this mesh.
	Material Material
}

// Objects represents the objects in a scene.
type Objects struct {
	// Meshes is a slice of all the meshes in the scene.
	Meshes []Mesh `yaml:"Meshes,omitempty"`
	// Triangles is a slice of all the triangles in the scene.
	Triangles []Triangle `yaml:"Triangles,omitempty"`
	// Spheres is a slice of all the spheres in the scene.
	Spheres []Sphere `yaml:"Spheres,omitempty"`
}

// Lights represents the lights in a scene.
type Lights struct {
	// Meshes is a slice of all the light meshes in the scene.
	Meshes []Mesh `yaml:"Meshes,omitempty"`
	// Triangles is a slice of all the light triangles in the scene.
	Triangles []Triangle `yaml:"Triangles,omitempty"`
	// Spheres is a slice of all the light spheres in the scene.
	Spheres []Sphere `yaml:"Spheres,omitempty"`
}

// Scene represents a scene that can be rendered.
type Scene struct {
	// Name is the name of the scene.
	Name string
	// Camera is the camera for this scene.
	Camera Camera
	// Objects contains all the objects in the scene.
	Objects Objects
	// Lights contains all the lights in the scene.
	Lights Lights
}
