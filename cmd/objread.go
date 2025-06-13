package main

import (
	"os"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func stringsToFloatSlice(parts []string) ([]float64, error) {
	floats := make([]float64, 0, len(parts))

	for _, s := range parts {
		if s == "" {
			continue
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		floats = append(floats, f)
	}

	return floats, nil
}

// fucking go
func stringsToIntSlice(parts []string) ([]int, error) {
	ints := make([]int, 0, len(parts))

	for _, s := range parts {
		if s == "" {
			continue
		}
		i, err := strconv.ParseInt(s, 0, 0)
		if err != nil {
			return nil, err
		}
		ints = append(ints, int(i))
	}

	return ints, nil
}

func loadObjFile(path string) (faces []Face) {
	data, err := os.ReadFile(path)
	check(err)
	return getObjData(string(data))
}

type Face struct {
	vertices  []Float3
	texCoords []Float2
	normals   []Float3
}

func (f Face) convertToTriangles() (vertices []Float3, vertexTexCoords []Float2, vertexNormals []Float3) {
	n := len(f.vertices)
	if n < 3 {
		panic("Not enough vertices in face for a triangle!")
	}

	vertices = make([]Float3, 0, f.getNumTriangles())
	vertexTexCoords = make([]Float2, 0, f.getNumTriangles())
	vertexNormals = make([]Float3, 0, f.getNumTriangles())

	// triangle fan
	for i := 1; i < n-1; i++ {
		vertices = append(vertices, f.vertices[0], f.vertices[i], f.vertices[i+1])
		vertexTexCoords = append(vertexTexCoords, f.texCoords[0], f.texCoords[i], f.texCoords[i+1])
		vertexNormals = append(vertexNormals, f.normals[0], f.normals[i], f.normals[i+1])
	}

	return
}

// ugh
func getObjData(objString string) (faces []Face) {
	allVertices := make([]Float3, 0)      // geometry vertices
	allTexturePoints := make([]Float2, 0) // texture atlas points
	allNormalPoints := make([]Float3, 0)  // face normals

	faces = make([]Face, 0) // output faces (face index groups)

	for line := range strings.SplitSeq(objString, "\n") {
		if len(line) > 1 {
			if line[:2] == "v " { // vertex positions
				axes, _ := stringsToFloatSlice(strings.Split(line[2:], " "))
				allVertices = append(allVertices, Float3{axes[0], axes[1], axes[2]})

			} else if line[:3] == "vt " { // texture data
				tpoint, _ := stringsToFloatSlice(strings.Split(line[3:], " "))
				allTexturePoints = append(allTexturePoints, Float2{tpoint[0], tpoint[1]})

			} else if line[:3] == "vn " { // normal data
				tpoint, _ := stringsToFloatSlice(strings.Split(line[3:], " "))
				allNormalPoints = append(allNormalPoints, Float3{tpoint[0], tpoint[1], tpoint[2]})

			} else if line[:2] == "f " { // face indices
				faceIndexGroups := strings.Split(line[2:], " ")

				face := Face{
					vertices:  make([]Float3, 0),
					texCoords: make([]Float2, 0),
					normals:   make([]Float3, 0),
				}

				// iterate over the groups
				for i := range faceIndexGroups {
					indexGroup, _ := stringsToIntSlice(strings.Split(faceIndexGroups[i], "/"))

					face.vertices = append(face.vertices, allVertices[indexGroup[0]-1])
					face.texCoords = append(face.texCoords, allTexturePoints[indexGroup[1]-1])
					face.normals = append(face.normals, allNormalPoints[indexGroup[2]-1])
				}

				faces = append(faces, face)
			}
		}
	}
	return
}
