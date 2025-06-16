package raster

import (
	"os"

	"golang.org/x/image/bmp"
)

// turn a bmp image into a colorbuffer
func BMPToImage(path string) (image Image) {
	// load the bmp
	file, err := os.Open(path)
	Check(err)
	defer file.Close()

	img, err := bmp.Decode(file)
	Check(err)

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
	pixelColor(Float2, Float3, float64) Float3
}

// texture shader

type TextureShader struct {
	texture Image
}

func (t TextureShader) pixelColor(coord Float2, _ Float3, _ float64) Float3 {
	return t.texture.sample(coord)
}

// lit shader

type LitShader struct {
	Color            Float3
	DirectionToLight Float3
}

func (l LitShader) pixelColor(_ Float2, normal Float3, _ float64) Float3 {
	normal = normal.normalized()
	lightIntensity := (dot3(normal, l.DirectionToLight) + 1) * 0.5
	return l.Color.mulscal(lightIntensity)
}

// litTexture shader, who could've seen it coming?

type LitTextureShader struct {
	Texture          Image
	DirectionToLight Float3
}

func (lt LitTextureShader) pixelColor(coord Float2, normal Float3, _ float64) Float3 {
	normal = normal.normalized()
	lightIntensity := (dot3(normal, lt.DirectionToLight) + 1) * 0.5
	return lt.Texture.sample(coord).mulscal(lightIntensity)
}

// terrain shader

type TerrainShader struct {
	DirectionToLight Float3
	Colors           []Float3
	Heights          []float64
	BGcol            Float3
}

var defaultHeights = []float64{
	0, 1.5, 6,
}

var defaultColors = []Float3{
	{0.2, 0.6, 0.98},   // water
	{0.2, 0.6, 0.1},    // grass
	{0.5, 0.35, 0.3},   // mountain
	{0.93, 0.93, 0.91}, // snow
}

func GetDefaultTerrainShaderHeights() []float64 {
	return defaultHeights
}

func GetDefaultTerrainShaderColors() []Float3 {
	return defaultColors
}

func (t TerrainShader) pixelColor(coord Float2, normal Float3, depth float64) Float3 {
	if t.Colors == nil {
		t.Colors = defaultColors
	}
	if t.Heights == nil {
		t.Heights = defaultHeights
	}

	normal = normal.normalized()
	lightIntensity := (dot3(normal, t.DirectionToLight) + 1) * 0.5
	triangleHeight := coord.X
	terrainCol := t.Colors[0]

	for i := range t.Heights {
		if triangleHeight > t.Heights[i] {
			terrainCol = t.Colors[i+1]
		} else {
			break
		}
	}

	noSee := 40.0
	fogBegins := 25.0
	var fogPercent float64

	if depth <= fogBegins {
		fogPercent = 0.0
	} else if depth >= noSee {
		fogPercent = 1.0
	} else {
		fogPercent = (depth - fogBegins) / (noSee - fogBegins)
	}

	return lerp(terrainCol.mulscal(lightIntensity), t.BGcol, fogPercent)
}
