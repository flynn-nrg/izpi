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
	"time"

	"github.com/flynn-nrg/izpi/internal/displacement"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"

	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"

	log "github.com/sirupsen/logrus"
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
	OBJ_FACE_TYPE_POLYGON
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
	Kd        []float32
	Ka        []float32
	Ks        []float32
	Ns        float32
	Ni        float32
	Sharpness int64
	D         float32
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
	var numFaces int64
	var currentGroup *Group
	var activeMaterial string
	var ignoreMaterials bool

	log.Info("Reading Wavefront data")
	startTime := time.Now()

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
			// Some programs export meshes without groups.
			// Create a default one.
			if currentGroup == nil {
				currentGroup = &Group{
					Name: "default",
				}
			}
			numFaces++
			f := strings.Split(s, " ")
			currentGroup.FaceType = OBJ_FACE_TYPE_POLYGON
			face, err := parseFace(f[1:])
			if err != nil {
				return nil, err
			}
			currentGroup.Faces = append(currentGroup.Faces, face)
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

	log.Infof("Parsed %v faces in %v", numFaces, time.Since(startTime))

	return o, nil
}

func (wo *WavefrontObj) NumGroups() int {
	return len(wo.Groups)
}

func (wo *WavefrontObj) GroupToTransportTrianglesWithMaterial(index int, materialName string) ([]*pb_transport.Triangle, error) {
	g := wo.Groups[index]
	switch g.FaceType {
	case OBJ_FACE_TYPE_POLYGON:
		triangles := []*pb_transport.Triangle{}
		for _, face := range g.Faces {
			triangles = append(triangles, wo.faceToTransportTriangle(face, materialName))
		}
		return triangles, nil
	default:
		return nil, ErrUnsupportedPolygonType
	}
}

func (wo *WavefrontObj) faceToTransportTriangle(face *Face, materialName string) *pb_transport.Triangle {

	vertex0 := wo.Vertices[face.Vertices[0].VIdx-1]
	vertex1 := wo.Vertices[face.Vertices[1].VIdx-1]
	vertex2 := wo.Vertices[face.Vertices[2].VIdx-1]

	return &pb_transport.Triangle{
		MaterialName: materialName,
		Vertex0: &pb_transport.Vec3{
			X: float32(vertex0.X),
			Y: float32(vertex0.Y),
			Z: float32(vertex0.Z),
		},
		Vertex1: &pb_transport.Vec3{
			X: float32(vertex1.X),
			Y: float32(vertex1.Y),
			Z: float32(vertex1.Z),
		},
		Vertex2: &pb_transport.Vec3{
			X: float32(vertex2.X),
			Y: float32(vertex2.Y),
			Z: float32(vertex2.Z),
		},
		Uv0: &pb_transport.Vec2{
			U: float32(wo.VertexUV[face.Vertices[0].VtIdx-1].U),
			V: float32(wo.VertexUV[face.Vertices[0].VtIdx-1].V),
		},
		Uv1: &pb_transport.Vec2{
			U: float32(wo.VertexUV[face.Vertices[1].VtIdx-1].U),
			V: float32(wo.VertexUV[face.Vertices[1].VtIdx-1].V),
		},
		Uv2: &pb_transport.Vec2{
			U: float32(wo.VertexUV[face.Vertices[2].VtIdx-1].U),
			V: float32(wo.VertexUV[face.Vertices[2].VtIdx-1].V),
		},
	}
}

// GroupToHitablesWithCustomMaterial returns a slice of hitables with the supplied material.
func (wo *WavefrontObj) GroupToHitablesWithCustomMaterial(index int, mat material.Material) ([]hitable.Hitable, error) {
	g := wo.Groups[index]
	switch g.FaceType {
	case OBJ_FACE_TYPE_POLYGON:
		return wo.groupToTrianglesWithCustomMaterial(g, mat)
	default:
		return nil, ErrUnsupportedPolygonType
	}
}

func (wo *WavefrontObj) groupToTrianglesWithCustomMaterial(g *Group, mat material.Material) ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}
	for _, face := range g.Faces {
		tris := wo.triangulate(face, mat)
		for _, tri := range tris {
			hitables = append(hitables, tri)
		}
	}

	return hitables, nil
}

// GroupToHitablesWithCustomMaterialAndDisplacement returns a slice of displaced hitables with the supplied material.
func (wo *WavefrontObj) GroupToHitablesWithCustomMaterialAndDisplacement(index int, mat material.Material, displacementMap texture.Texture, min, max float32) ([]hitable.Hitable, error) {
	g := wo.Groups[index]
	switch g.FaceType {
	case OBJ_FACE_TYPE_POLYGON:
		return wo.groupToTrianglesWithCustomMaterialAndDisplacement(g, mat, displacementMap, min, max)
	default:
		return nil, ErrUnsupportedPolygonType
	}
}

func (wo *WavefrontObj) groupToTrianglesWithCustomMaterialAndDisplacement(g *Group, mat material.Material, displacementMap texture.Texture, min, max float32) ([]hitable.Hitable, error) {
	hitables := []hitable.Hitable{}
	for _, face := range g.Faces {
		tris := wo.triangulate(face, mat)
		displaced, err := displacement.ApplyDisplacementMap(tris, displacementMap, min, max)
		if err != nil {
			return hitables, err
		}
		for _, tri := range displaced {
			hitables = append(hitables, tri)
		}
	}

	return hitables, nil
}

func (wo *WavefrontObj) triangulate(face *Face, mat material.Material) []*hitable.Triangle {
	switch len(face.Vertices) {
	case 3:
		vertex0 := wo.Vertices[face.Vertices[0].VIdx-1]
		vertex1 := wo.Vertices[face.Vertices[1].VIdx-1]
		vertex2 := wo.Vertices[face.Vertices[2].VIdx-1]
		uv0 := &texture.UV{}
		uv1 := &texture.UV{}
		uv2 := &texture.UV{}

		if wo.HasUV && !wo.IgnoreTextures {
			uv0 = wo.VertexUV[face.Vertices[0].VtIdx-1]
			uv1 = wo.VertexUV[face.Vertices[1].VtIdx-1]
			uv2 = wo.VertexUV[face.Vertices[2].VtIdx-1]
		}
		if wo.IgnoreNormals {
			return []*hitable.Triangle{hitable.NewTriangleWithUV(
				vertex0, vertex1, vertex2, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat)}
		} else {
			if face.Vertices[0].VnIdx != face.Vertices[1].VnIdx || face.Vertices[0].VnIdx != face.Vertices[2].VnIdx {
				vn0 := wo.VertexNormals[face.Vertices[0].VnIdx-1]
				vn1 := wo.VertexNormals[face.Vertices[1].VnIdx-1]
				vn2 := wo.VertexNormals[face.Vertices[2].VnIdx-1]
				return []*hitable.Triangle{hitable.NewTriangleWithUVAndVertexNormals(
					vertex0, vertex1, vertex2, vn0, vn1, vn2, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat)}
			} else {
				vn0 := wo.VertexNormals[face.Vertices[0].VnIdx-1]
				return []*hitable.Triangle{hitable.NewTriangleWithUVAndNormal(
					vertex0, vertex1, vertex2, vn0, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat)}
			}
		}

	case 4:
		vertex0 := wo.Vertices[face.Vertices[0].VIdx-1]
		vertex1 := wo.Vertices[face.Vertices[1].VIdx-1]
		vertex2 := wo.Vertices[face.Vertices[2].VIdx-1]
		vertex3 := wo.Vertices[face.Vertices[3].VIdx-1]
		if wo.HasUV && !wo.IgnoreTextures {
			uv0 := wo.VertexUV[face.Vertices[0].VtIdx-1]
			uv1 := wo.VertexUV[face.Vertices[1].VtIdx-1]
			uv2 := wo.VertexUV[face.Vertices[2].VtIdx-1]
			uv3 := wo.VertexUV[face.Vertices[3].VtIdx-1]
			return []*hitable.Triangle{
				hitable.NewTriangleWithUV(vertex0, vertex1, vertex2, uv0.U, uv0.V, uv1.U, uv1.V, uv2.U, uv2.V, mat),
				hitable.NewTriangleWithUV(vertex0, vertex2, vertex3, uv0.U, uv0.V, uv2.U, uv2.V, uv3.U, uv3.V, mat),
			}
		} else {
			return []*hitable.Triangle{
				hitable.NewTriangle(vertex0, vertex1, vertex2, mat),
				hitable.NewTriangle(vertex0, vertex2, vertex3, mat),
			}
		}
	}

	// TODO: Implement this for real so that any polygon is split up in triangles.
	log.Warnf("Unsupported face type with %v vertices", len(face.Vertices))
	return []*hitable.Triangle{}
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

func parseFloats(s []string) ([]float32, error) {
	res := []float32{}
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
