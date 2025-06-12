package main

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// -------------------------- rasterizer stuff

// rasterizer information
type SWR struct {
	running bool
	surface *sdl.Surface
	window  *sdl.Window

	initiated bool

	buffer *sdl.Surface
	pbuf   image

	// other data
	models []model
}

// initiating the rasterizer
func (s *SWR) init(window *sdl.Window) {
	// set the gamestate
	s.running = true

	// load the game window
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	s.surface = surface

	// set window
	s.window = window
}

// -------------------------- helper funcs

// get the slice of pixels of a surface. remember to lock and unlock!!
func getPixels(surface *sdl.Surface) []uint32 {
	pixels := (*[1 << 30]uint32)(surface.Data())[:surface.W*surface.H] // cast address to a pointer
	return pixels
}

// create an empty color buff
func newColorBuffer(w, h int, color float3) (image [][]float3) {
	image = make([][]float3, h)
	for y := range image {
		image[y] = make([]float3, w)
		for x := range image[y] {
			image[y][x] = color
		}
	}
	return
}

// create an empty depth buff
func newDepthBuffer(w, h int) (image [][]float64) {
	image = make([][]float64, h)
	for y := range image {
		image[y] = make([]float64, w)
		for x := range image[y] {
			image[y][x] = math.Inf(1)
		}
	}
	return
}

// -------------------------- update

// core update for the software renderer
func (s *SWR) update() {
	// ---------------------- init
	if !s.initiated {
		// set our buffer up
		var err error
		s.buffer, err = sdl.CreateRGBSurface(0, s.surface.W, s.surface.H, 32, 0x00FF0000, 0x0000FF00, 0x000000FF, 0xFF000000)
		if err != nil {
			panic(err)
		}
		// set out pixel buffer - buffer up
		s.pbuf = newImage(int(s.buffer.W), int(s.buffer.W))
		// clear it
		s.buffer.FillRect(nil, 0)

		// load in a model
		s.models = make([]model, 1)
		s.models[0].points = append(s.models[0].points, loadObjFile("assets/suzy.obj")...)
		// model colors
		s.models[0].triangleCols = make([]float3, len(s.models[0].points)/3)
		for i := range s.models[0].triangleCols {
			s.models[0].triangleCols[i] = randomColor()
			//fmt.Printf("%v, %v, %v\n", s.models[0].triangleCols[i].x, s.models[0].triangleCols[i].y, s.models[0].triangleCols[i].z)
		}
		s.models[0].transform.position.z = 10

		s.initiated = true
	}

	// ---------------------- logic

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			println("Quit")
			s.running = false
			return
		}
	}

	// controls
	keys := sdl.GetKeyboardState()
	// rot
	if keys[sdl.SCANCODE_UP] != 0 {
		s.models[0].transform.pitch += rotSpeed
	}
	if keys[sdl.SCANCODE_DOWN] != 0 {
		s.models[0].transform.pitch -= rotSpeed
	}
	if keys[sdl.SCANCODE_LEFT] != 0 {
		s.models[0].transform.yaw -= rotSpeed
	}
	if keys[sdl.SCANCODE_RIGHT] != 0 {
		s.models[0].transform.yaw += rotSpeed
	}
	// pos
	if keys[sdl.SCANCODE_W] != 0 {
		s.models[0].transform.position.z += 0.05
	}
	if keys[sdl.SCANCODE_S] != 0 {
		s.models[0].transform.position.z -= 0.05
	}
	if keys[sdl.SCANCODE_Q] != 0 {
		s.models[0].transform.position.y += 0.05
	}
	if keys[sdl.SCANCODE_E] != 0 {
		s.models[0].transform.position.y -= 0.05
	}
	if keys[sdl.SCANCODE_A] != 0 {
		s.models[0].transform.position.x -= 0.05
	}
	if keys[sdl.SCANCODE_D] != 0 {
		s.models[0].transform.position.x += 0.05
	}

	// // spinny
	// for i := range s.models {
	// 	s.models[i].transform.yaw += 0.05
	// 	s.models[i].transform.pitch += 0.05
	// }

	// --------------------- drawing

	// clear
	s.surface.FillRect(nil, 0)
	s.pbuf.colorBuffer = newColorBuffer(int(s.buffer.W), int(s.buffer.H), float3{0, 0, 0})
	s.pbuf.depthBuffer = newDepthBuffer(int(s.buffer.W), int(s.buffer.H))

	// -------------- draw

	// manipulate s.pbuf
	for _, model := range s.models {
		s.pbuf = render(s.pbuf, int(s.buffer.W), int(s.buffer.H), model.transform, model.points, model.triangleCols, 60*math.Pi/180.0)
	}

	// get the pixels on the buffer
	pixels := getPixels(s.buffer)

	// load the image buffer into the surface buffer
	for y, row := range s.pbuf.colorBuffer {
		for x := range row {
			pixels[y*int(s.buffer.W)+x] = s.pbuf.colorBuffer[y][x].toUint32()
			// fmt.Printf("(%v, %v, %v), %08x\n", s.pbuf[y][x].x, s.pbuf[y][x].y, s.pbuf[y][x].z, s.pbuf[y][x].toUint32())
		}
	}

	// write the buffer to the window
	s.buffer.Blit(nil, s.surface, &sdl.Rect{X: 0, Y: 0})

	// update window
	s.window.UpdateSurface()

	sdl.Delay(33)
}
