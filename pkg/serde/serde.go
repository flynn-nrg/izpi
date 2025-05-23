// Package serde implements functions to serialise and deserialise scene data.
package serde

import (
	"io"
)

const (
	// ConstantTexture represents a constant texture.
	ConstantTexture = "constant"
	// ImageTexture represents an image texture.
	ImageTexture = "image"
	// NoiseTexture represents a noise texture.
	NoiseTexture = "noise"
	// LambertMaterial represents a Lambertian material.
	LambertMaterial = "lambert"
	// DielectricMaterial represents a dielectric material.
	DielectricMaterial = "dielectric"
	// DiffuseLightMaterial represents a diffuse light material.
	DiffuseLightMaterial = "diffuse_light"
	// IsotropicMaterial represents an isotropic material.
	IsotropicMaterial = "isotropic"
	// MetalMaterial represents a metallic material.
	MetalMaterial = "metal"
	// PBRMaterial represents a physically based rendering material.
	PBRMaterial = "pbr"
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
	X float64 `yaml:"x"`
	// Y is the y coordinate of this vector.
	Y float64 `yaml:"y"`
	// Z is the z coordinate of this vector.
	Z float64 `yaml:"z"`
}

// Camera represents a camera.
type Camera struct {
	// LookFrom is the location of the camera.
	LookFrom Vec3 `yaml:"lookfrom"`
	// LookAt is where the camera is pointing at.
	LookAt Vec3 `yaml:"lookat"`
	// VUp defines the "up" vector.
	VUp Vec3 `yaml:"vup"`
	// VFov define the field of view.
	VFov float64 `yaml:"vfov"`
	// Aspect is the aspect ratio.
	Aspect float64 `yaml:"aspect,omitempty"`
	// Aperture is the lens aperture for this camera.
	Aperture float64 `yaml:"aperture"`
	// FocusDist is the focus distance.
	FocusDist float64 `yaml:"focusdist"`
	// Time0 defines the beginning time of the exposure.
	Time0 float64 `yaml:"time0"`
	// Time1 defines the end time of the exposure.
	Time1 float64 `yaml:"time1"`
}

// Image represents an image based texture.
type Image struct {
	// FileName is the name of the file containing the texture data.
	FileName string `yaml:"filename"`
}

// Constant represents a constant texture.
type Constant struct {
	// Value is a vector with the constant data.
	Value Vec3 `yaml:"value"`
}

// Noise represents a Perlin noise texture.
type Noise struct {
	// Scale defines the Perlin noise scale.
	Scale float64 `yaml:"scale"`
}

// Texture represents a texture instance.
type Texture struct {
	// Name is the name of the texture.
	Name string `yaml:"name"`
	// Type is the texture type: constant, image or noise.
	Type string `yaml:"type"`
	// Image is an instance of an Image texture.
	Image Image `yaml:"image,omitempty"`
	// Constant is an instance of a Constant texture.
	Constant Constant `yaml:"constant,omitempty"`
	// Noise is an instance of a Perlin noise texture.
	Noise Noise `yaml:"noise,omitempty"`
}

// Lambert represents a Lambertian material.
type Lambert struct {
	// Albedo is the colour texture.
	Albedo Texture `yaml:"albedo"`
}

// DiffuseLight represents a diffuse light material.
type DiffuseLight struct {
	// Emit is the colour texture.
	Emit Texture `yaml:"emit"`
}

// Isotropic represents a dielectric material.
type Isotropic struct {
	// Albedo is the colour texture.
	Albedo Texture `yaml:"albedo"`
}

// Metal represents a metallic material.
type Metal struct {
	// Albedo is the colour texture.
	Albedo Vec3 `yaml:"albedo"`
	// Fuzz defines how shiny a metallic surface is. 0 is a perfect mirror.
	Fuzz float64 `yaml:"fuzz"`
}

// Dielectric represents a dielectric material.
type Dielectric struct {
	// RefIdx is the refraction index.
	RefIdx float64 `yaml:"refidx"`
}

// PBR represents a physically based rendering material.
type PBR struct {
	// Albedo is the colour texture.
	Albedo Texture `yaml:"albedo"`
	// NormalMap is the normal map texture.
	NormalMap *Texture `yaml:"normalMap,omitempty"`
	// Roughness is the roughness texture.
	Roughness Texture `yaml:"roughness"`
	// Metalness is the metalness texture.
	Metalness Texture `yaml:"metalness"`
	// SSS is the subsurface scattering strength texture.
	SSS Texture `yaml:"sss"`
	// SSSRadius is the subsurface scattering radius.
	SSSRadius float64 `yaml:"sssRadius"`
}

// Material represents a single material.
type Material struct {
	// Name is the name of the material.
	Name string `yaml:"name"`
	// Type is the type of material: lambert, diffuse_light, isotropic, metal, dielectric, pbr.
	Type string `yaml:"type"`
	// Lambert is a lambert material.
	Lambert Lambert `yaml:"lambert,omitempty"`
	// DiffuseLight is a diffuse light.
	DiffuseLight DiffuseLight `yaml:"diffuselight,omitempty"`
	// Isotropic is an isotropic material.
	Isotropic Isotropic `yaml:"isotropic,omitempty"`
	// Metal is a metallic material.
	Metal Metal `yaml:"metal,omitempty"`
	// Dielectric is a dielectric material.
	Dielectric Dielectric `yaml:"dielectric,omitempty"`
	// PBR is a physically based rendering material.
	PBR PBR `yaml:"pbr,omitempty"`
}

// Displacement represents a displacement mapping operator.
type Displacement struct {
	// DisplacementMap is the displacement map that gets applied.
	DisplacementMap Image `yaml:"displacementmap,omitempty"`
	// Min is the lower value of the displacement.
	Min float64 `yaml:"min"`
	// Max is the upper value of the displacement.
	Max float64 `yaml:"max"`
}

// Sphere represents a sphere.
type Sphere struct {
	// Center defines the centre of the sphere.
	Center Vec3 `yaml:"center"`
	// Radius is the radius of the sphere.
	Radius float64 `yaml:"radius"`
	// Material is the material of the sphere.
	Material Material `yaml:"material"`
}

// Triangle represents a single triangle.
// Vertices must be defined counter clockwise.
type Triangle struct {
	// Vertex0 is the first vertex of this triangle.
	Vertex0 Vec3 `yaml:"vertex0"`
	// Vertex1 is the second vertex of this triangle.
	Vertex1 Vec3 `yaml:"vertex1"`
	// Vertex2 is the third vertex of this triangle.
	Vertex2 Vec3 `yaml:"vertex2"`
	// U0 is the u coordinate of the first vertex.
	U0 float64 `yaml:"u0,omitempty"`
	// V0 is the v coordinate of the first vertex.
	V0 float64 `yaml:"v0,omitempty"`
	// U1 is the u coordinate of the second vertex.
	U1 float64 `yaml:"u1,omitempty"`
	// V1 is the v coordinate of the second vertex.
	V1 float64 `yaml:"v1,omitempty"`
	// U2 is the u coordinate of the third vertex.
	U2 float64 `yaml:"u2,omitempty"`
	// V2 is the v coordinate of the third vertex.
	V2 float64 `yaml:"v2,omitempty"`
	// Displacement is the displacement map associated with this triangle.
	Displacement Displacement `yaml:"displacement,omitempty"`
	// Material is the material for this triangle.
	Material Material `yaml:"material"`
	// Type is the type of triangle.
	Type string `yaml:"type"`
	// FlipNormals indicates whether the normals should be flipped.
	FlipNormals bool `yaml:"flipNormals,omitempty"`
	// Min is the minimum point for this triangle.
	Min Vec3 `yaml:"min,omitempty"`
	// Max is the maximum point for this triangle.
	Max Vec3 `yaml:"max,omitempty"`
	// K is the k value for this triangle.
	K float64 `yaml:"k,omitempty"`
}

// Mesh represents a Wavefront OBJ instance.
type Mesh struct {
	// WavefrontData is the name of the file containing the mesh information.
	WavefrontData string `yaml:"wavefrontdata"`
	// Translate is the translation vector that is applied to all the vertices.
	Translate Vec3 `yaml:"translate"`
	// Scale is a the scale vector that is applied to all the vetices.
	Scale Vec3 `yaml:"scale"`
	// Displacement is the displacement map associated with this triangle.
	Displacement Displacement `yaml:"displacement,omitempty"`
	// Material is the material associated with this mesh.
	Material Material `yaml:"material"`
}

// Objects represents the objects in a scene.
type Objects struct {
	// Meshes is a slice of all the meshes in the scene.
	Meshes []Mesh `yaml:"meshes,omitempty"`
	// Triangles is a slice of all the triangles in the scene.
	Triangles []Triangle `yaml:"triangles,omitempty"`
	// Spheres is a slice of all the spheres in the scene.
	Spheres []Sphere `yaml:"spheres,omitempty"`
}

// Scene represents a scene that can be rendered.
type Scene struct {
	// Name is the name of the scene.
	Name string `yaml:"name"`
	// Camera is the camera for this scene.
	Camera Camera `yaml:"camera"`
	// Objects contains all the objects in the scene.
	Objects Objects `yaml:"objects"`
}
