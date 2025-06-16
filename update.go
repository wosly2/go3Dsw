package raster

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

// -------------------------- rasterizer stuff

// rasterizer information
type SoftwareRasterizer struct {
	Running bool
	surface *sdl.Surface
	Window  *sdl.Window

	initiated bool

	Buffer     *sdl.Surface // the buffer that is drawn to the screen.
	MetaBuffer Image        // meta buffer is rendered to by the rasterizer. it is then rendered to SoftwareRasterizer.Buffer.

	Process UpdateProcess
}

// initiating the rasterizer
func (s *SoftwareRasterizer) Init(window *sdl.Window) {
	// set the gamestate
	s.Running = true

	// load the game window
	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}
	s.surface = surface

	// set window
	s.Window = window
}

// -------------------------- helper funcs

// get the slice of pixels of a surface. remember to lock and unlock!!
func getPixels(surface *sdl.Surface) []uint32 {
	pixels := (*[1 << 30]uint32)(surface.Data())[:surface.W*surface.H] // cast address to a pointer
	return pixels
}

// create an empty color buff
func NewColorBuffer(w, h int, color Float3) (image [][]Float3) {
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
func NewDepthBuffer(w, h int) (image [][]float64) {
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

func (s *SoftwareRasterizer) Run(scene *Scene, process UpdateProcess) {
	s.Process = process
	for s.Running {
		s.Update(scene)
	}
}

func (s *SoftwareRasterizer) Stop() {
	s.Running = false
}

var lastTime uint64 = sdl.GetTicks64()
var fps float64 = 0

func FPS() float64 {
	return fps
}

type UpdateProcess interface {
	Update(s *SoftwareRasterizer, scene *Scene)
}

// core Update for the software renderer
func (s *SoftwareRasterizer) Update(scene *Scene) {
	// ---------------------- init
	if !s.initiated {
		// set our buffer up
		var err error
		s.Buffer, err = sdl.CreateRGBSurface(0, s.surface.W, s.surface.H, 32, 0x00FF0000, 0x0000FF00, 0x000000FF, 0xFF000000)
		if err != nil {
			panic(err)
		}
		// set out pixel buffer - buffer up
		s.MetaBuffer = newImage(int(s.Buffer.W), int(s.Buffer.H))
		// clear it
		s.Buffer.FillRect(nil, 0)

		s.initiated = true

		scene.Cam.transform.UpdateBases()
	}

	// ---------------------- fps

	currentTime := sdl.GetTicks64()
	delta := currentTime - lastTime
	lastTime = currentTime

	if delta > 0 {
		fps = 1000.0 / float64(delta)
	}

	// ---------------------- logic

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			println("Quit")
			s.Running = false
			return
		}
	}

	// controls
	keys := sdl.GetKeyboardState()
	// rot
	if keys[sdl.SCANCODE_UP] != 0 {
		scene.Cam.transform.Pitch += rotSpeed
		scene.Cam.transform.Pitch = clamp(scene.Cam.transform.Pitch, ToRadians(-85), ToRadians(85))
		scene.Cam.transform.UpdateBases()
	}
	if keys[sdl.SCANCODE_DOWN] != 0 {
		scene.Cam.transform.Pitch -= rotSpeed
		scene.Cam.transform.Pitch = clamp(scene.Cam.transform.Pitch, ToRadians(-85), ToRadians(85))
		scene.Cam.transform.UpdateBases()
	}
	if keys[sdl.SCANCODE_LEFT] != 0 {
		scene.Cam.transform.Yaw += rotSpeed
		scene.Cam.transform.UpdateBases()
	}
	if keys[sdl.SCANCODE_RIGHT] != 0 {
		scene.Cam.transform.Yaw -= rotSpeed
		scene.Cam.transform.UpdateBases()
	}
	// pos
	// get bases
	ihat, _, khat := scene.Cam.transform.GetBasisVectors()
	if keys[sdl.SCANCODE_W] != 0 {
		scene.Cam.transform.Position = scene.Cam.transform.Position.add(khat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_S] != 0 {
		scene.Cam.transform.Position = scene.Cam.transform.Position.sub(khat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_Q] != 0 {
		scene.Cam.transform.Position.Y -= moveSpeed
	}
	if keys[sdl.SCANCODE_E] != 0 {
		scene.Cam.transform.Position.Y += moveSpeed
	}
	if keys[sdl.SCANCODE_A] != 0 {
		scene.Cam.transform.Position = scene.Cam.transform.Position.sub(ihat.mulscal(moveSpeed))
	}
	if keys[sdl.SCANCODE_D] != 0 {
		scene.Cam.transform.Position = scene.Cam.transform.Position.add(ihat.mulscal(moveSpeed))
	}

	// --------------------- drawing

	// clear
	s.MetaBuffer.fillcb(scene.BGcol)
	s.MetaBuffer.filldb()

	// -------------- draw
	s.MetaBuffer = renderScene(Scene(*scene), s.MetaBuffer)

	// get the pixels on the buffer
	pixels := getPixels(s.Buffer)

	// load the image buffer into the surface buffer, but upside down for funzies
	for y, row := range s.MetaBuffer.colorBuffer {
		for x := range row {
			pixels[y*int(s.Buffer.W)+x] = s.MetaBuffer.colorBuffer[s.MetaBuffer.h-y-1][x].toUint32()
			// fmt.Printf("(%v, %v, %v), %08x\n", s.pbuf[y][x].x, s.pbuf[y][x].y, s.pbuf[y][x].z, s.pbuf[y][x].toUint32())
		}
	}

	if s.Process != nil { // update the user process
		UpdateProcess(s.Process).Update(s, scene)
	}

	// write the buffer to the window
	s.Buffer.Blit(nil, s.surface, &sdl.Rect{X: 0, Y: 0})

	// update window
	s.Window.UpdateSurface()

	sdl.Delay(33)
}
