package main

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
	raster "github.com/wosly2/go3Dsw"
	"github.com/wosly2/go3Dsw/font"
)

type Process struct { // implements raster.UpdateProcess
	font font.Font
}

func (p Process) Update(swr *raster.SoftwareRasterizer, sc *raster.Scene) {
	// gui
	var helloWorld *sdl.Surface = p.font.RenderString(fmt.Sprintf("fps: %v", math.Round(raster.FPS())), 1, 0, 1)
	helloWorld.BlitScaled(nil, swr.Buffer, &sdl.Rect{X: 5, Y: 5, W: helloWorld.W * 3, H: helloWorld.H * 3})
}

func main() {
	// init sdl
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// create window
	window, err := sdl.CreateWindow("Software Rasterizer", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 1400, 800, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// init renderer
	swr := raster.SoftwareRasterizer{}
	swr.Init(window)
	raster.Check(err)

	// init scene
	mainScene := raster.NewScene()

	//mainScene.bgCol = Float3{0.396, 0.773, 1}
	mainScene.BGcol = raster.Float3{X: 1, Y: 1, Z: 1}

	// // chunker
	// mainScene.cam.transform.position.y = 5
	// mainScene.chunker = &Chunker{shader: TerrainShader{
	// 	Float3{0, 1, 0},
	// 	defaultColors,
	// 	defaultHeights,
	// 	mainScene.bgCol,
	// }, resolution: 30, chunkSize: 20}

	mainScene.AddModel(raster.NewModel(raster.ModelInitOptions{
		ID:           "suzy",
		LoadFromPath: true,
		Path:         "../assets/suzy.obj",
		GiveColors:   false,
	}))
	mainScene.GetModel("suzy").Transform.Position = raster.Float3{X: 0, Y: 0, Z: 8}
	mainScene.GetModel("suzy").Transform.Pitch = raster.ToRadians(-90)
	mainScene.GetModel("suzy").Transform.Scale = raster.Float3{X: 2, Y: 2, Z: 2}
	mainScene.GetModel("suzy").Transform.UpdateBases()
	mainScene.GetModel("suzy").Shader = raster.LitShader{
		Color:            raster.Float3{X: 0.396, Y: 0.773, Z: 1},
		DirectionToLight: raster.Float3{X: 0, Y: -0.5, Z: -1},
	}

	process := Process{}

	process.font = font.MakeDefaultFont()

	swr.Run(&mainScene, process)

	defer swr.Window.Destroy()
}
