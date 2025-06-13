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
		id:           "suzy",
		loadFromPath: true,
		path:         "assets/suzy.obj",
		randColors:   true,
	}))
	mainScene.getModel("suzy").transform.position = Float3{0, 0, 8}
	mainScene.getModel("suzy").transform.yaw = toRadians(180)
	mainScene.getModel("suzy").transform.scale = Float3{1, 1, 1}
	mainScene.getModel("suzy").shader = LitTextureShader{
		texture:          bmpToImage("assets/c1.bmp"),
		directionToLight: Float3{1, 1, 1},
	}

	for swr.running {
		swr.update(&mainScene)
	}
	defer swr.window.Destroy()
}
