package main

import (
	"fmt"
	"math"
)

// ------------ SCENE -------------

type Scene struct {
	models map[string]*Model
	cam    Camera
}

func newScene() (s Scene) {
	s.models = make(map[string]*Model, 0)
	s.cam.fov = defaultFov
	return
}

func (s *Scene) addModel(model *Model) {
	// ensure the id is unique
	_, ok := s.models[model.id]
	if ok {
		fmt.Println("\033[31mModel ID not unique, was not loaded to scene.\033[0m")
		return
	}

	// set it
	s.models[model.id] = model
}

func (s *Scene) getModel(id string) *Model {
	model, ok := s.models[id]
	if !ok {
		panic(fmt.Sprintf("No such model %v!", id))
	}
	return model
}

func renderScene(s Scene, target Image) (image Image) {
	image = target
	for _, model := range s.models {
		image = render(image, Model(*model), s.cam)
	}
	return
}

// ------------ MODEL -------------

type Model struct {
	id           string
	faces        []Face
	transform    Transform
	shader       Shader
	triangleCols []Float3
}

type Transform struct {
	yaw      float64
	pitch    float64
	position Float3
}

func (t Transform) toWorldPoint(p Float3) Float3 {
	ihat, jhat, khat := t.getBasisVectors()
	return transformVector(ihat, jhat, khat, p).add(t.position)
}

func (t Transform) toLocalPoint(worldPoint Float3) Float3 {
	ihat, jhat, khat := t.getInverseBasisVectors()
	return transformVector(ihat, jhat, khat, worldPoint.sub(t.position))
}

func (t Transform) getInverseBasisVectors() (ihat, jhat, khat Float3) {
	// ---- Yaw ----
	ihat_yaw := Float3{math.Cos(-t.yaw), 0, math.Sin(-t.yaw)}
	jhat_yaw := Float3{0, 1, 0}
	khat_yaw := Float3{-math.Sin(-t.yaw), 0, math.Cos(-t.yaw)}
	// ---- Pitch ----
	ihat_pitch := Float3{1, 0, 0}
	jhat_pitch := Float3{0, math.Cos(-t.pitch), -math.Sin(-t.pitch)}
	khat_pitch := Float3{0, math.Sin(-t.pitch), math.Cos(-t.pitch)}
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
	ihat_yaw := Float3{math.Cos(t.yaw), 0, math.Sin(t.yaw)}
	jhat_yaw := Float3{0, 1, 0}
	khat_yaw := Float3{-math.Sin(t.yaw), 0, math.Cos(t.yaw)}
	// ---- Pitch ----
	ihat_pitch := Float3{1, 0, 0}
	jhat_pitch := Float3{0, math.Cos(t.pitch), -math.Sin(t.pitch)}
	khat_pitch := Float3{0, math.Sin(t.pitch), math.Cos(t.pitch)}
	// ---- Yaw and Pitch combined ----
	ihat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, ihat_pitch)
	jhat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, jhat_pitch)
	khat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, khat_pitch)

	return
}

type ModelInitOptions struct {
	// important data
	id string

	// flags and shit
	loadFromPath   bool
	path           string
	loadFromPoints bool
	points         []Float3
	giveColors     bool
	colors         []Float3
	randColors     bool
}

func (f Face) getNumTriangles() int {
	return len(f.vertices) - 2
}

func (m Model) getNumTriangles() (n int) {
	for _, face := range m.faces {
		n += face.getNumTriangles()
	}
	return
}

func newModel(o ModelInitOptions) *Model {
	// init the model
	var model Model
	model.transform = Transform{0, 0, Float3{0, 0, 0}}

	model.id = o.id

	// load in the points
	if o.loadFromPoints { // deprecated
		panic("Blub")
	} else if o.loadFromPath {
		model.faces = loadObjFile(o.path)
	} else {
		panic("No source point data configured while loading model!")
	}

	// load in the colors
	if o.randColors { // deprecated
		model.triangleCols = make([]Float3, model.getNumTriangles())
		for i := range model.triangleCols {
			model.triangleCols[i] = randomColor()
		}
	} else if o.giveColors {
		model.triangleCols = o.colors
	} else {
		panic("No source color data configured while loading model!")
	}

	return &model
}
