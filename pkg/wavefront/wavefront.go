// Package wavefront implements functions to parse Wavefront OBJ files and transform the data.
// This is not meant to be a fully compliant parser but rather something good enough that allows for easy use
// of triangle meshes.
package wavefront

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"gitlab.com/flynn-nrg/izpi/pkg/hitable"
	"gitlab.com/flynn-nrg/izpi/pkg/material"
	"gitlab.com/flynn-nrg/izpi/pkg/texture"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

type ParseOption int

const (
	INVALID_OPTION ParseOption = iota
	IGNORE_NORMALS
	IGNORE_MATERIALS
	IGNORE_TEXTURES
)

type ObjFaceType int

const (
	OBJ_FACE_TYPE_INVALID ObjFaceType = iota
	OBJ_FACE_TYPE_TRIANGLE
	OBJ_FACE_TYPE_QUAD
)

var (
	ErrUnsupportedPolygonType = errors.New("unsupported polygon type")
	ErrInvalidFaceData        = errors.New("invalid face data")
)

// WavefrontObj represents the data contained in a Wavefront OBJ file.
type WavefrontObj struct {
	IgnoreMaterials bool
	IgnoreNormals   bool
	IgnoreTextures  bool
	HasNormals      bool
	HasUV           bool
	Centre          *vec3.Vec3Impl
	ObjectName      string
	Vertices        []*vec3.Vec3Impl
	VertexNormals   []*vec3.Vec3Impl
	VertexUV        []*texture.UV
	MtlLib          map[string]*Material
	Groups          []*Group
}

// Material represents a material defined in a .mtl file.
type Material struct {
	Name      string
	Kd        []float64
	Ka        []float64
	Ks        []float64
	Ns        float64
	Ni        float64
	Sharpness int64
	D         float64
	Illum     int64
}

// VertexIndices represents the indices used by a vertex.
type VertexIndices struct {
	VIdx  int64
	VtIdx int64
	VnIdx int64
}

// Face represents a single face.
type Face struct {
	Vertices []*VertexIndices
}

// Group represents a single group.
type Group struct {
	Name     string
	FaceType ObjFaceType
	Material string
	Faces    []*Face
}

// NewObjFromReader returns WavefrontObj with the geometry contained in a Wavefront .obj file.
func NewObjFromReader(r io.Reader, containerDirectory string, opts ...ParseOption) (*WavefrontObj, error) {
	var currentGroup *Group
	var activeMaterial string
	var ignoreMaterials bool

	o := &WavefrontObj{
		// Objects are expected to have their centre at the origin.
		Centre: &vec3.Vec3Impl{},
	}

	for _, option := range opts {
		switch option {
		case IGNORE_MATERIALS:
			ignoreMaterials = true
			o.IgnoreMaterials = true
		case IGNORE_NORMALS:
			o.IgnoreNormals = true
		case IGNORE_TEXTURES:
			o.IgnoreTextures = true
		}
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "#") {
			continue
		}
		if strings.HasPrefix(s, "o") {
			objectName := strings.Split(s, " ")
			if len(objectName) == 2 {
				o.ObjectName = objectName[1]
			}
			continue
		}
		if strings.HasPrefix(s, "v ") {
			values := strings.Split(s, " ")
			v, err := parseFloats(values[1:])
			if err != nil {
				return nil, err
			}
			o.Vertices = append(o.Vertices, &vec3.Vec3Impl{X: v[0], Y: v[1], Z: v[2]})
			continue
		}
		if strings.HasPrefix(s, "vn") {
			o.HasNormals = true
			values := strings.Split(s, " ")
			vn, err := parseFloats(values[1:])
			if err != nil {
				return nil, err
			}
			o.VertexNormals = append(o.VertexNormals, &vec3.Vec3Impl{X: vn[0], Y: vn[1], Z: vn[2]})
			continue
		}
		if strings.HasPrefix(s, "vt") {
			o.HasUV = true
			values := strings.Split(s, " ")
			vt, err := parseFloats(values[1:])
			if err != nil {
				return nil, err
			}
			o.VertexUV = append(o.VertexUV, &texture.UV{U: vt[0], V: vt[1]})
			continue
		}
		if strings.HasPrefix(s, "f") {
			f := strings.Split(s, " ")
			switch len(f) {
			case 4:
				currentGroup.FaceType = OBJ_FACE_TYPE_TRIANGLE
				face, err := parseFace(f[1:])
				if err != nil {
					return nil, err
				}
				currentGroup.Faces = append(currentGroup.Faces, face)
			default:
				return nil, ErrUnsupportedPolygonType
			}

			continue
		}
		if strings.HasPrefix(s, "mtllib") {
			if ignoreMaterials {
				continue
			}
			mtllibStr := strings.Split(s, " ")
			fileName := fmt.Sprintf("%v/%v", containerDirectory, mtllibStr[1])
			mtlLib, err := parseMaterialFile(fileName)
			if err != nil {
				return nil, err
			}
			o.MtlLib = mtlLib
			continue
		}
		if strings.HasPrefix(s, "usemtl") {
			material := strings.Split(s, " ")
			activeMaterial = material[1]
			continue
		}
		if strings.HasPrefix(s, "g") {
			if currentGroup != nil {
				o.Groups = append(o.Groups, currentGroup)
			}
			name := strings.Split(s, " ")
			currentGroup = &Group{
				Name:     name[1],
				Material: activeMaterial,
			}
			continue
		}
	}

	// Add pending group
	o.Groups = append(o.Groups, currentGroup)

	return o, nil
}

func (wo *WavefrontObj) NumGroups() int {
	return len(wo.Groups)
}

// GroupToHitablesWithCustomMaterial returns a slice of hitables with the supplied material.
func (wo *WavefrontObj) GroupToHitablesWithCustomMaterial(index int, mat material.Material) ([]hitable.Hitable, error) {
	g := wo.Groups[index]
	switch g.FaceType {
	case OBJ_FACE_TYPE_TRIANGLE:
		return wo.groupToTrianglesWithCustomMaterial(g, mat)
	default:
		return nil, ErrUnsupportedPolygonType
	}
}

func (wo *WavefrontObj) groupToTrianglesWithCustomMaterial(g *Group, mat material.Material) ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}
	for _, face := range g.Faces {
		vertex0 := wo.Vertices[face.Vertices[0].VIdx-1]
		vertex1 := wo.Vertices[face.Vertices[1].VIdx-1]
		vertex2 := wo.Vertices[face.Vertices[2].VIdx-1]
		if wo.HasUV && !wo.IgnoreTextures {
			uv0 := wo.VertexUV[face.Vertices[0].VtIdx-1]
			uv1 := wo.VertexUV[face.Vertices[1].VtIdx-1]
			uv2 := wo.VertexUV[face.Vertices[2].VtIdx-1]
			if wo.IgnoreNormals {
				hitables = append(hitables, hitable.NewTriangleWithUV(
					vertex0, vertex1, vertex2, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat))
			} else {
				normal := wo.VertexNormals[face.Vertices[0].VnIdx-1]
				hitables = append(hitables, hitable.NewTriangleWithUVAndNormal(
					vertex0, vertex1, vertex2, normal, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat))
			}
		} else {
			hitables = append(hitables, hitable.NewTriangle(vertex0, vertex1, vertex2, mat))
		}
	}

	return hitables, nil
}

// Translate translates all the vertices in this object by the specified amount.
func (wo *WavefrontObj) Translate(translate *vec3.Vec3Impl) {
	wo.Centre = vec3.Add(wo.Centre, translate)
	for _, v := range wo.Vertices {
		v.X += translate.X
		v.Y += translate.Y
		v.Z += translate.Z
	}
}

// Scale scales all the vertices in this object by the specified amount.
func (wo *WavefrontObj) Scale(scale *vec3.Vec3Impl) {
	for _, v := range wo.Vertices {
		v.X = ((v.X - wo.Centre.X) * scale.X) + wo.Centre.X
		v.Y = ((v.Y - wo.Centre.Y) * scale.Y) + wo.Centre.Y
		v.Z = ((v.Z - wo.Centre.Z) * scale.Z) + wo.Centre.Z
	}
}

func parseFloats(s []string) ([]float64, error) {
	res := []float64{}
	for _, i := range s {
		f, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return res, err
		}
		res = append(res, f)
	}
	return res, nil
}

func parseFace(s []string) (*Face, error) {
	face := &Face{}
	for _, v := range s {
		fv, err := parseFaceVertex(v)
		if err != nil {
			return nil, err
		}
		face.Vertices = append(face.Vertices, fv)
	}

	return face, nil
}

// parseFaceVertex expects an vertex formatted like 1/1/1
func parseFaceVertex(s string) (*VertexIndices, error) {
	var idx []int64

	indices := strings.Split(s, "/")
	if len(indices) != 3 {
		return nil, ErrInvalidFaceData
	}

	for _, index := range indices {
		// Some fields might not exist and that's ok.
		i, err := strconv.ParseInt(index, 10, 32)
		if err != nil {
			i = 0
		}
		idx = append(idx, i)
	}

	return &VertexIndices{
		VIdx:  idx[0],
		VtIdx: idx[1],
		VnIdx: idx[2],
	}, nil
}

// parseMaterialFile parses a .mtl file. This code assumes machine-generated well-formatted
// and does not make any attempt at recovering from a broken file.
func parseMaterialFile(fileName string) (map[string]*Material, error) {
	var currentMaterial *Material

	mtlLib := make(map[string]*Material)

	r, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "#") {
			continue
		}
		if strings.HasPrefix(s, "newmtl") {
			if currentMaterial != nil {
				mtlLib[currentMaterial.Name] = currentMaterial
			}

			name := strings.Split(s, " ")
			currentMaterial = &Material{
				Name: name[1],
			}

			continue
		}
		if strings.HasPrefix(s, "Kd") {
			kdStr := strings.Split(s, " ")
			kd, err := parseFloats(kdStr[1:])
			if err != nil {
				return nil, err
			}
			currentMaterial.Kd = kd
			continue
		}
		if strings.HasPrefix(s, "Ns") {
			nsStr := strings.Split(s, " ")
			ns, err := strconv.ParseFloat(nsStr[1], 32)
			if err != nil {
				return nil, err
			}
			currentMaterial.Ns = ns
			continue
		}
		if strings.HasPrefix(s, "Ni") {
			niStr := strings.Split(s, " ")
			ni, err := strconv.ParseFloat(niStr[1], 32)
			if err != nil {
				return nil, err
			}
			currentMaterial.Ni = ni
			continue
		}
		if strings.HasPrefix(s, "d") {
			dStr := strings.Split(s, " ")
			d, err := strconv.ParseFloat(dStr[1], 32)
			if err != nil {
				return nil, err
			}
			currentMaterial.D = d
			continue
		}
		if strings.HasPrefix(s, "illum") {
			illumStr := strings.Split(s, " ")
			illum, err := strconv.ParseInt(illumStr[1], 10, 32)
			if err != nil {
				return nil, err
			}
			currentMaterial.Illum = illum
		}
		if strings.HasPrefix(s, "Ka") {
			kaStr := strings.Split(s, " ")
			ka, err := parseFloats(kaStr[1:])
			if err != nil {
				return nil, err
			}
			currentMaterial.Ka = ka
			continue
		}
		if strings.HasPrefix(s, "Ks") {
			ksStr := strings.Split(s, " ")
			ks, err := parseFloats(ksStr[1:])
			if err != nil {
				return nil, err
			}
			currentMaterial.Ks = ks
		}
	}

	mtlLib[currentMaterial.Name] = currentMaterial

	return mtlLib, nil
}
