package main

import (
	"os"

	"golang.org/x/image/bmp"
)

// turn a bmp image into a colorbuffer
func bmpToImage(path string) (image Image) {
	// load the bmp
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	img, err := bmp.Decode(file)
	check(err)

	// extract the colors
	image = newImage(img.Bounds().Dx(), img.Bounds().Dy())
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			image.colorBuffer[y][x] = Float3{float64(r), float64(g), float64(b)}.mulscal(1.0 / 255 / 255)
		}
	}

	return
}

// ----------- Shaders ------------

type Shader interface {
	pixelColor(Float2) Float3
}

// texture shader

type TextureShader struct {
	texture Image
}

func (t TextureShader) pixelColor(coord Float2) Float3 {
	return t.texture.sample(coord)
}
