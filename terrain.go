package raster

import (
	"fmt"
	"math"
	"math/rand/v2"

	"github.com/KEINOS/go-noise"
)

const seed int64 = 42

var noiseGen, _ = noise.New(noise.OpenSimplex, seed)

func calculateElevation(pos Float2) float64 {
	const layerCount int = 5
	const lacunarity float64 = 2      // how fast detail increases per layer
	const persistence float64 = 0.5   // how quickly strength of layers decrease
	const ridgeLayerStart float64 = 3 // make noise more ridge like later on

	frequency := 0.05
	amplitude := 1.0
	elevation := 0.0

	for i := range layerCount {
		n := noiseGen.Eval64(pos.X*frequency, pos.Y*frequency)
		if i >= int(ridgeLayerStart) {
			n = 0.5 - math.Abs(n)
		}
		elevation += n * amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return elevation * 10
}

func generatePointMap(resolution int, worldSize float64, gridCenter Float2) (pointMap [][]Float3) {
	// make the pointmap
	pointMap = make([][]Float3, resolution)
	for i := range pointMap {
		pointMap[i] = make([]Float3, resolution)
	}

	for y := range resolution {
		for x := range resolution {
			localGridPos_sNorm := Float2{float64(x), float64(y)}.mulscal(1.0 / (float64(resolution) - 1.0)).sub(Float2{0.5, 0.5})
			gridWorldPos := gridCenter.add(localGridPos_sNorm.mulscal(worldSize))
			gridWorldPos = calculateJiggle(gridWorldPos)
			elevation := max(0, calculateElevation(gridWorldPos)+1.8)
			pointMap[y][x] = Float3{gridWorldPos.X, elevation, gridWorldPos.Y}
		}
	}

	return
}

func calculateJiggle(v Float2) Float2 {
	const jiggleStrength float64 = 0.05
	return Float2{
		v.X + noiseGen.Eval64(v.X+1000, v.Y+1000)*jiggleStrength,
		v.Y + noiseGen.Eval64(v.X-1000, v.Y-1000)*jiggleStrength,
	}
}

func generateTerrain(resolution int, worldsize float64, gridCenter Float2, shader Shader, name ...string) *Model {
	pointMap := generatePointMap(resolution, worldsize, gridCenter)
	faces := make([]Face, 0)

	for y := range resolution - 1 {
		for x := range resolution - 1 {
			a, b, c, d := pointMap[y][x], pointMap[y+1][x], pointMap[y][x+1], pointMap[y+1][x+1]

			// texture coordinates
			t1 := Float2{a.Y + b.Y + c.Y, 0.0}.mulscal(1.0 / 3.0)
			t2 := Float2{b.Y + d.Y + c.Y, 0.0}.mulscal(1.0 / 3.0)

			// normals
			n1 := cross3(b.sub(a), c.sub(b)).normalized() // tri 1
			n2 := cross3(d.sub(b), c.sub(d)).normalized() // tri 2

			faces = append(faces, Face{
				vertices:  []Float3{a, b, c},
				texCoords: []Float2{t1, t1, t1},
				normals:   []Float3{n1, n1, n1},
			})

			faces = append(faces, Face{
				vertices:  []Float3{b, d, c},
				texCoords: []Float2{t2, t2, t2},
				normals:   []Float3{n2, n2, n2},
			})

		}
	}

	// assign a random id
	var id string
	if len(name) >= 1 {
		id = name[0]
	} else {
		id = fmt.Sprintf("chunk%x", rand.Uint64())
	}

	return &Model{
		ID:        id,
		Faces:     faces,
		Transform: Transform{Position: Float3{0, 0, 10}, Scale: Float3{1, 1, 1}},
		Shader:    shader,
	}
}

// chunking

type Chunker struct {
	terrainChunksActive []Model
	terrainChunkLookup  map[[2]int]Model
	shader              Shader

	resolution int
	chunkSize  float64
}

func (c *Chunker) updateTerrainChunks(camPos Float3, resolution int, chunkSize float64) {
	if c.terrainChunkLookup == nil {
		c.terrainChunkLookup = make(map[[2]int]Model)
	}

	centerX := int(math.Round(camPos.X / chunkSize))
	centerY := int(math.Round(camPos.Z / chunkSize))
	c.terrainChunksActive = make([]Model, 0) // clear

	// create a 5x5 grid of terrain chunks centered around camera position
	for y := centerY - 2; y <= centerY+2; y++ {
		for x := centerX - 2; x <= centerX+2; x++ {
			chunk, ok := c.terrainChunkLookup[[2]int{x, y}]
			if !ok {
				center := Float2{float64(x) * chunkSize, float64(y) * chunkSize} // chunk center in world sapce
				chunk = *generateTerrain(resolution, chunkSize, center, c.shader)
				chunk.Transform.UpdateBases()
			}

			c.terrainChunksActive = append(c.terrainChunksActive, chunk) // add to draw list
			c.terrainChunkLookup[[2]int{x, y}] = chunk                   // add to the lookup table
		}
	}
}
