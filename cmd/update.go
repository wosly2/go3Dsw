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
	pbuf   Image
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
func newColorBuffer(w, h int, color Float3) (image [][]Float3) {
	image = make([][]Float3, h)
	for y := range image {
		image[y] = make([]Float3, w)
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
func (s *SWR) update(scene *Scene) {
	// ---------------------- init
	if !s.initiated {
		// set our buffer up
		var err error
		s.buffer, err = sdl.CreateRGBSurface(0, s.surface.W, s.surface.H, 32, 0x00FF0000, 0x0000FF00, 0x000000FF, 0xFF000000)
		if err != nil {
			panic(err)
		}
		// set out pixel buffer - buffer up
		s.pbuf = newImage(int(s.buffer.W), int(s.buffer.H))
		// clear it
		s.buffer.FillRect(nil, 0)

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
		scene.cam.transform.pitch -= rotSpeed
		scene.cam.transform.pitch = clamp(scene.cam.transform.pitch, toRadians(-85), toRadians(85))
	}
	if keys[sdl.SCANCODE_DOWN] != 0 {
		scene.cam.transform.pitch += rotSpeed
	}
	if keys[sdl.SCANCODE_LEFT] != 0 {
		scene.cam.transform.yaw += rotSpeed
	}
	if keys[sdl.SCANCODE_RIGHT] != 0 {
		scene.cam.transform.yaw -= rotSpeed
	}
	// pos
	// get bases
	ihat, jhat, khat := scene.cam.transform.getBasisVectors()
	if keys[sdl.SCANCODE_W] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.add(khat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_S] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.sub(khat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_Q] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.add(jhat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_E] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.sub(jhat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_A] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.sub(ihat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_D] != 0 {
		scene.cam.transform.position = scene.cam.transform.position.add(ihat.mulscal(moveSpeed))
	}

	// --------------------- drawing

	// clear
	s.surface.FillRect(nil, 0)
	s.pbuf.colorBuffer = newColorBuffer(int(s.buffer.W), int(s.buffer.H), Float3{0, 0, 0})
	s.pbuf.depthBuffer = newDepthBuffer(int(s.buffer.W), int(s.buffer.H))

	// -------------- draw
	s.pbuf = renderScene(Scene(*scene), s.pbuf)

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
