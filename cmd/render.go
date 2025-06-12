package main

import (
	"math"
)

// -------------------------- render

func vertexToScreen(vertex float3, trans transform, numPixels float2, fov float64) float3 {
	vertex_world := trans.toWorldPoint(vertex)

	var screenHeight_World float64 = math.Tan(fov / 2)
	pixelsPerWorldUnit := numPixels.y / screenHeight_World / vertex_world.z

	pixelOffset := float2{vertex_world.x * pixelsPerWorldUnit, vertex_world.y * pixelsPerWorldUnit}
	vertex_screen := pixelOffset.add(float2{numPixels.x / 2, numPixels.y / 2})
	return float3{vertex_screen.x, vertex_screen.y, vertex_world.z}
}

func render(img image, w, h int, trans transform, points []float3, triangleCols []float3, fov float64) image {
	fw, fh := float64(w), float64(h)

	for i := 0; i < len(points); i += 3 {
		// fmt.Println(i / 3)
		a := vertexToScreen(points[i+0], trans, float2{fw, fh}, fov)
		b := vertexToScreen(points[i+1], trans, float2{fw, fh}, fov)
		c := vertexToScreen(points[i+2], trans, float2{fw, fh}, fov)
		// triangle bounds
		minX := min(min(a.x, b.x), c.x)
		minY := min(min(a.y, b.y), c.y)
		maxX := max(max(a.x, b.x), c.x)
		maxY := max(max(a.y, b.y), c.y)
		// pixel block covering bounds
		blockStartX := clamp(minX, 0, fw-1)
		blockStartY := clamp(minY, 0, fh-1)
		blockEndX := clamp(maxX, 0, fw-1)
		blockEndY := clamp(maxY, 0, fh-1)

		for y := int(blockStartY); y <= int(blockEndY); y++ {
			for x := int(blockStartX); x <= int(blockEndX); x++ {
				p := float2{float64(x), float64(y)}
				inTri, weights := pointInTriangle(a.make2(), b.make2(), c.make2(), p)
				if inTri {
					// sum the weighted depths at each vertex and get depth for the current pixel
					depths := float3{a.z, b.z, c.z}
					depth := dot3(depths, weights)
					if depth > img.depthBuffer[y][x] {
						continue
					}
					// update pixel otherwise
					img.colorBuffer[y][x] = triangleCols[i/3]
					img.depthBuffer[y][x] = depth
				}
			}
		}
	}

	return img
}

// image type
type image struct {
	colorBuffer [][]float3
	depthBuffer [][]float64
	w           int
	h           int
}

func (i image) fs() float2 {
	return float2{float64(i.w), float64(i.h)}
}

func newImage(x, y int) (img image) {
	img = image{
		colorBuffer: make([][]float3, y),
		depthBuffer: make([][]float64, y),
		w:           x,
		h:           y,
	}
	// fill buffers
	for i := range y {
		img.colorBuffer[i] = make([]float3, x)
		img.depthBuffer[i] = make([]float64, x)
	}
	return
}
