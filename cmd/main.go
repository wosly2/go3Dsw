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

	for swr.running {
		swr.update()
	}
	defer swr.window.Destroy()
}
