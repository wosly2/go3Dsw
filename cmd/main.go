package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	// init sdl
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	// create window
	window, err := sdl.CreateWindow("Software Rasterizer", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	// init renderer
	swr := SWR{}
	swr.init(window)

	// init scene
	mainScene := newScene()

	mainScene.addModel(newModel(ModelInitOptions{
		id:           "cube",
		loadFromPath: true,
		path:         "assets/cube.obj",
		randColors:   true,
	}))
	mainScene.getModel("cube").transform.position = Float3{0, 0, 8}
	mainScene.getModel("cube").shader = TextureShader{
		texture: bmpToImage("assets/p2.bmp"),
	}

	for swr.running {
		swr.update(&mainScene)
	}
	defer swr.window.Destroy()
}
