package raster

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
	depth := vertex_view.Z

	var screenHeight_World float64 = math.Tan(cam.fov / 2)
	pixelsPerWorldUnit := numPixels.Y / screenHeight_World / depth

	pixelOffset := Float2{vertex_view.X * pixelsPerWorldUnit, vertex_view.Y * pixelsPerWorldUnit}
	vertex_screen := pixelOffset.add(Float2{numPixels.X / 2, numPixels.Y / 2})
	return Float3{vertex_screen.X, vertex_screen.Y, vertex_view.Z}
}

func render(img Image, model Model, cam Camera) Image {

	for _, face := range model.Faces {
		triangleVertices, vertexTexCoords, vertexNormals := face.convertToTriangles()
		for i := 0; i < len(triangleVertices); i += 3 {
			// println(model.faceCount, curFace, model.vertices[curFace], trisDrawnThisFace)
			a := vertexToScreen(triangleVertices[i+0], model.Transform, cam, img.fs())
			b := vertexToScreen(triangleVertices[i+1], model.Transform, cam, img.fs())
			c := vertexToScreen(triangleVertices[i+2], model.Transform, cam, img.fs())

			if a.Z <= 0 || b.Z <= 0 || c.Z <= 0 { // skip tri if vertex is behind cam
				continue
			}

			// triangle bounds
			minX := min(min(a.X, b.X), c.X)
			minY := min(min(a.Y, b.Y), c.Y)
			maxX := max(max(a.X, b.X), c.X)
			maxY := max(max(a.Y, b.Y), c.Y)

			// pixel block covering bounds
			blockStartX := clamp(minX, 0, img.fs().X-1)
			blockStartY := clamp(minY, 0, img.fs().Y-1)
			blockEndX := clamp(maxX, 0, img.fs().X-1)
			blockEndY := clamp(maxY, 0, img.fs().Y-1)

			for y := int(blockStartY); y <= int(blockEndY); y++ {
				for x := int(blockStartX); x <= int(blockEndX); x++ {
					p := Float2{float64(x), float64(y)}
					inTri, weights := pointInTriangle(a.make2(), b.make2(), c.make2(), p)
					if inTri {
						// depth check
						depths := Float3{a.Z, b.Z, c.Z}
						depth := 1 / dot3(depths.under(1), weights)
						if depth > img.depthBuffer[y][x] {
							continue
						}

						// update pixel otherwise
						if model.Shader == nil {
							panic(fmt.Sprintf("No shader selected on model %v!", model.ID))
						}

						// texture weighting
						var texCoord Float2
						texCoord = texCoord.add(vertexTexCoords[i+0].mulscal(1 / depths.X).mulscal(weights.X))
						texCoord = texCoord.add(vertexTexCoords[i+1].mulscal(1 / depths.Y).mulscal(weights.Y))
						texCoord = texCoord.add(vertexTexCoords[i+2].mulscal(1 / depths.Z).mulscal(weights.Z))
						texCoord = texCoord.mulscal(depth)

						// normal weighting
						var normal Float3
						normal = normal.add(vertexNormals[i+0].mulscal(1 / depths.X).mulscal(weights.X))
						normal = normal.add(vertexNormals[i+1].mulscal(1 / depths.Y).mulscal(weights.Y))
						normal = normal.add(vertexNormals[i+2].mulscal(1 / depths.Z).mulscal(weights.Z))
						normal = normal.mulscal(depth)
						//ihat, jhat, khat := model.transform.GetInverseBasisVectors()
						//rotatedNormal := ihat.mulscal(normal.x).add(jhat.mulscal(normal.y)).add(khat.mulscal(normal.z))

						img.colorBuffer[y][x] = model.Shader.pixelColor(texCoord, normal, depth) //model.triangleCols[i/3].mulscal(0.1).add(model.shader.pixelColor(texCoord, normal).mulscal(0.9))
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

func (i *Image) fillcb(color Float3) {
	for y := range i.h {
		for x := range i.w {
			i.colorBuffer[y][x] = color
		}
	}
}

func (i *Image) filldb(depth ...float64) {
	var d float64
	if len(depth) >= 1 {
		d = depth[0]
	} else {
		d = math.MaxFloat32
	}
	for y := range i.h {
		for x := range i.w {
			i.depthBuffer[y][x] = d
		}
	}
}

func (i Image) sample(coord Float2) Float3 {
	// clamp coord to [0, 1]
	coord.X = clamp(coord.X, 0, 1)
	coord.Y = clamp(coord.Y, 0, 1)

	// calculate nearest texel
	x := int((coord.X) * (i.fs().X))
	y := int((coord.Y) * (i.fs().Y)) // had a bug here where it was coord.x and it fucked me over for forever

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
