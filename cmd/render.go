package main

import (
	"fmt"
	"math"
)

// -------------------------- render

type Camera struct {
	fov       float64
	transform Transform
}

func vertexToScreen(vertex Float3, transform Transform, cam Camera, numPixels Float2) Float3 {
	vertex_world := transform.toWorldPoint(vertex)
	vertex_view := cam.transform.toLocalPoint(vertex_world)

	var screenHeight_World float64 = math.Tan(cam.fov / 2)
	pixelsPerWorldUnit := numPixels.y / screenHeight_World / vertex_view.z

	pixelOffset := Float2{vertex_view.x * pixelsPerWorldUnit, vertex_view.y * pixelsPerWorldUnit}
	vertex_screen := pixelOffset.add(Float2{numPixels.x / 2, numPixels.y / 2})
	return Float3{vertex_screen.x, vertex_screen.y, vertex_view.z}
}

func render(img Image, model Model, cam Camera) Image {

	for _, face := range model.faces {
		triangleVertices, vertexTexCoords, _ := face.convertToTriangles()
		for i := 0; i < len(triangleVertices); i += 3 {
			// println(model.faceCount, curFace, model.vertices[curFace], trisDrawnThisFace)
			a := vertexToScreen(triangleVertices[i+0], model.transform, cam, img.fs())
			b := vertexToScreen(triangleVertices[i+1], model.transform, cam, img.fs())
			c := vertexToScreen(triangleVertices[i+2], model.transform, cam, img.fs())

			if a.z <= 0 || b.z <= 0 || c.z <= 0 { // skip tri if vertex is behind cam
				continue
			}

			// triangle bounds
			minX := min(min(a.x, b.x), c.x)
			minY := min(min(a.y, b.y), c.y)
			maxX := max(max(a.x, b.x), c.x)
			maxY := max(max(a.y, b.y), c.y)

			// pixel block covering bounds
			blockStartX := clamp(minX, 0, img.fs().x-1)
			blockStartY := clamp(minY, 0, img.fs().y-1)
			blockEndX := clamp(maxX, 0, img.fs().x-1)
			blockEndY := clamp(maxY, 0, img.fs().y-1)

			for y := int(blockStartY); y <= int(blockEndY); y++ {
				for x := int(blockStartX); x <= int(blockEndX); x++ {
					p := Float2{float64(x), float64(y)}
					inTri, weights := pointInTriangle(a.make2(), b.make2(), c.make2(), p)
					if inTri {
						// depth check
						depths := Float3{a.z, b.z, c.z}
						depth := 1 / dot3(depths.under(1), weights)
						if depth > img.depthBuffer[y][x] {
							continue
						}

						// update pixel otherwise
						if model.shader == nil {
							panic(fmt.Sprintf("No shader selected on model %v!", model.id))
						}

						// FIXME: texture mapping
						var texCoord Float2
						// p :=
						texCoord = texCoord.add(vertexTexCoords[i+0].mulscal(weights.x / depths.x))
						texCoord = texCoord.add(vertexTexCoords[i+1].mulscal(weights.y / depths.y))
						texCoord = texCoord.add(vertexTexCoords[i+2].mulscal(weights.z / depths.z))
						texCoord = texCoord.mulscal(depth)

						img.colorBuffer[y][x] = model.triangleCols[i/3].mulscal(0.1).add(model.shader.pixelColor(texCoord).mulscal(0.9))
						//println(texCoord.x, texCoord.y)
						img.depthBuffer[y][x] = depth
					}
				}
			}
		}
	}

	return img
}

// Image type
type Image struct {
	colorBuffer [][]Float3
	depthBuffer [][]float64
	w           int
	h           int
}

func (i Image) sample(coord Float2) Float3 {
	// clamp coord to [0, 1]
	coord.x = clamp(coord.x, 0, 1)
	coord.y = clamp(coord.y, 0, 1)

	// calculate nearest texel
	x := int((coord.x) * (i.fs().x))
	y := int((coord.y) * (i.fs().y)) // had a bug here where it was coord.x and it fucked me over for forever

	// clamp dos
	if x < 0 {
		x = 0
	} else if x >= i.w {
		x = i.w - 1
	}
	if y < 0 {
		y = 0
	} else if y >= i.h {
		y = i.h - 1
	}

	return i.colorBuffer[y][x]
}

func (i Image) fs() Float2 {
	return Float2{float64(i.w), float64(i.h)}
}

func newImage(x, y int) (img Image) {
	img = Image{
		colorBuffer: make([][]Float3, y),
		depthBuffer: make([][]float64, y),
		w:           x,
		h:           y,
	}
	// fill buffers
	for i := range y {
		img.colorBuffer[i] = make([]Float3, x)
		img.depthBuffer[i] = make([]float64, x)
	}
	return
}
