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

func loadObjFile(path string) []float3 {
	data, err := os.ReadFile(path)
	check(err)
	return getObjTriangles(string(data))
}

// shitty parser, apparently
func getObjTriangles(objString string) []float3 {
	allPoints := make([]float3, 0)
	trianglePoints := make([]float3, 0) // each set of three is a triangle

	for _, line := range strings.Split(objString, "\n") {
		if len(line) > 1 {
			if line[:2] == "v " { // vertex positions
				axes, _ := stringsToFloatSlice(strings.Split(line[2:], " "))
				allPoints = append(allPoints, float3{axes[0], axes[1], axes[2]})
			} else if line[:2] == "f " { // face indices
				faceIndexGroups := strings.Split(line[2:], " ")
				for i := range faceIndexGroups {
					indexGroup, _ := stringsToIntSlice(strings.Split(faceIndexGroups[i], "/"))
					pointIndex := indexGroup[0] - 1 // subtract one since indices start at 1 in obj

					if i >= 3 { // n-gon triangle fan
						trianglePoints = append(trianglePoints, trianglePoints[len(trianglePoints)-(3*i-6)])
						trianglePoints = append(trianglePoints, trianglePoints[len(trianglePoints)-2])
					}
					trianglePoints = append(trianglePoints, allPoints[pointIndex])
				}
			}
		}
	}
	return trianglePoints
}
