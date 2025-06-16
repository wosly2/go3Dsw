package raster

import (
	"fmt"
	"math"
)

// ------------ SCENE -------------

type Scene struct {
	Models  map[string]*Model
	Cam     Camera
	BGcol   Float3
	Chunker *Chunker
}

func NewScene() (s Scene) {
	s.Models = make(map[string]*Model, 0)
	s.Cam.fov = defaultFov
	s.Cam.transform.Scale = Float3{1, 1, 1}
	return
}

func (s *Scene) AddModel(model *Model) {
	// ensure the id is unique
	_, ok := s.Models[model.ID]
	if ok {
		fmt.Println("\033[31mModel ID not unique, was not loaded to scene.\033[0m")
		return
	}

	// set it
	s.Models[model.ID] = model
}

func (s *Scene) GetModel(id string) *Model {
	model, ok := s.Models[id]
	if !ok {
		panic(fmt.Sprintf("No such model %v!", id))
	}
	return model
}

func renderScene(s Scene, target Image) (image Image) {
	image = target
	for _, model := range s.Models {
		image = render(image, Model(*model), s.Cam)
	}
	if s.Chunker != nil {
		s.Chunker.updateTerrainChunks(s.Cam.transform.Position, s.Chunker.resolution, s.Chunker.chunkSize)
		for _, model := range s.Chunker.terrainChunksActive {
			image = render(image, model, s.Cam)
		}
	}

	return
}

// ------------ MODEL -------------

type Model struct {
	ID        string
	Faces     []Face
	Transform Transform
	Shader    Shader
}

type Transform struct {
	Yaw      float64
	Pitch    float64
	Position Float3
	Scale    Float3

	ihat, jhat, khat, ihat_inv, jhat_inv, khat_inv Float3
}

func (t *Transform) SetRotation(pitch, yaw float64) {
	t.Pitch = pitch
	t.Yaw = yaw

	t.UpdateBases()
}

func (t *Transform) UpdateBases() {
	t.ihat, t.jhat, t.khat = t.getBasisVectors()
	t.ihat_inv, t.jhat_inv, t.khat_inv = t.getInverseBasisVectors()
}

func (t *Transform) SetPitch(pitch float64) {
	t.SetRotation(pitch, t.Yaw)
}

func (t *Transform) SetYaw(yaw float64) {
	t.SetRotation(t.Pitch, yaw)
}

func (t Transform) GetInverseBasisVectors() (ihat, jhat, khat Float3) {
	//debugutil.Println("Got inverse basis from cache.")
	return t.ihat_inv, t.jhat_inv, t.khat_inv
}

func (t Transform) GetBasisVectors() (ihat, jhat, khat Float3) {
	//debugutil.Println("Got basis from cache.")
	return t.ihat, t.jhat, t.khat
}

func (t Transform) toWorldPoint(p Float3) Float3 {
	ihat, jhat, khat := t.GetBasisVectors()
	return transformVector(ihat.mulscal(t.Scale.X), jhat.mulscal(t.Scale.Y), khat.mulscal(t.Scale.Z), p).add(t.Position)
}

func (t Transform) toLocalPoint(worldPoint Float3) Float3 {
	ihat, jhat, khat := t.GetInverseBasisVectors()
	return transformVector(ihat.mulscal(1/t.Scale.X), jhat.mulscal(1/t.Scale.Y), khat.mulscal(1/t.Scale.Z), worldPoint.sub(t.Position))
}

func (t Transform) getInverseBasisVectors() (ihat, jhat, khat Float3) {
	// ---- Yaw ----
	ihat_yaw := Float3{math.Cos(-t.Yaw), 0, math.Sin(-t.Yaw)}
	jhat_yaw := Float3{0, 1, 0}
	khat_yaw := Float3{-math.Sin(-t.Yaw), 0, math.Cos(-t.Yaw)}
	// ---- Pitch ----
	ihat_pitch := Float3{1, 0, 0}
	jhat_pitch := Float3{0, math.Cos(-t.Pitch), -math.Sin(-t.Pitch)}
	khat_pitch := Float3{0, math.Sin(-t.Pitch), math.Cos(-t.Pitch)}
	// ---- Yaw and Pitch combined ----
	ihat = transformVector(ihat_pitch, jhat_pitch, khat_pitch, ihat_yaw)
	jhat = transformVector(ihat_pitch, jhat_pitch, khat_pitch, jhat_yaw)
	khat = transformVector(ihat_pitch, jhat_pitch, khat_pitch, khat_yaw)

	// FIXME: this code below should work but is giving me a black screen
	// ihatl, khatl, jhatl := t.getBasisVectors()
	// ihat = Float3{ihatl.x, jhatl.x, khatl.x}
	// jhat = Float3{ihatl.y, jhatl.y, khatl.y}
	// khat = Float3{ihatl.z, jhatl.z, khatl.z}

	return
}

func (t Transform) getBasisVectors() (ihat, jhat, khat Float3) {
	// ---- Yaw ----
	ihat_yaw := Float3{math.Cos(t.Yaw), 0, math.Sin(t.Yaw)}
	jhat_yaw := Float3{0, 1, 0}
	khat_yaw := Float3{-math.Sin(t.Yaw), 0, math.Cos(t.Yaw)}
	// ---- Pitch ----
	ihat_pitch := Float3{1, 0, 0}
	jhat_pitch := Float3{0, math.Cos(t.Pitch), -math.Sin(t.Pitch)}
	khat_pitch := Float3{0, math.Sin(t.Pitch), math.Cos(t.Pitch)}
	// ---- Yaw and Pitch combined ----
	ihat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, ihat_pitch)
	jhat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, jhat_pitch)
	khat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, khat_pitch)

	return
}

type ModelInitOptions struct {
	// important data
	ID string

	// flags and shit
	LoadFromPath   bool
	Path           string
	LoadFromPoints bool
	Points         []Float3
	GiveColors     bool
	Colors         []Float3
	RandColors     bool
}

func (f Face) getNumTriangles() int {
	return len(f.vertices) - 2
}

func (m Model) getNumTriangles() (n int) {
	for _, face := range m.Faces {
		n += face.getNumTriangles()
	}
	return
}

func NewModel(o ModelInitOptions) *Model {
	// init the model
	var model Model
	model.Transform = Transform{0, 0, Float3{0, 0, 0}, Float3{0, 0, 0}, Float3{}, Float3{}, Float3{}, Float3{}, Float3{}, Float3{}}

	model.ID = o.ID

	// load in the points
	if o.LoadFromPoints { // deprecated
		panic("Blub")
	} else if o.LoadFromPath {
		model.Faces = loadObjFile(o.Path)
	} else {
		panic("No source point data configured while loading model!")
	}

	// // load in the colors
	// if o.randColors { // deprecated
	// 	model.triangleCols = make([]Float3, model.getNumTriangles())
	// 	for i := range model.triangleCols {
	// 		model.triangleCols[i] = randomColor()
	// 	}
	// } else if o.giveColors {
	// 	model.triangleCols = o.colors
	// } else {
	// 	panic("No source color data configured while loading model!")
	// }

	return &model
}
