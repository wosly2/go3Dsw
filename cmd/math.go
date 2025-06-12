package main

import (
	"math"
	"math/rand"
)

type float3 struct {
	x, y, z float64
}

type float2 struct {
	x, y float64
}

func (v float2) add(other float2) float2 {
	return float2{v.x + other.x, v.y + other.y}
}

func (v float2) sub(other float2) float2 {
	return float2{v.x - other.x, v.y - other.y}
}

func (v float2) mulscal(scalar float64) float2 {
	return float2{v.x * scalar, v.y * scalar}
}

func (v float3) add(other float3) float3 {
	return float3{v.x + other.x, v.y + other.y, v.z + other.z}
}

func (v float3) sub(other float3) float3 {
	return float3{v.x - other.x, v.y - other.y, v.z - other.z}
}

func (v float3) mulscal(scalar float64) float3 {
	return float3{v.x * scalar, v.y * scalar, v.z * scalar}
}

func (v float3) make2() float2 {
	return float2{v.x, v.y}
}

// clamp a float64
func clamp(n, lo, hi float64) float64 {
	if n > hi {
		return hi
	}
	if n < lo {
		return lo
	}
	return n
}

// convert a float3 to 32-bit colorspace
func (f float3) toUint32() (u uint32) {
	// ensure they're clipped to 0-1
	r := uint32(clamp(f.x, 0, 1) * 255)
	g := uint32(clamp(f.y, 0, 1) * 255)
	b := uint32(clamp(f.z, 0, 1) * 255)
	a := uint32(255) // set default alpha

	return (a << 24) | (r << 16) | (g << 8) | b
}

// dot product of a and b
// equal to product of lengths * cosine of their angle
func dot(a, b float2) float64 {
	return a.x*b.x + a.y*b.y
}

// dot product of a and b, but float3!
func dot3(a, b float3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

// finds perpendicular vector (90 clockwise)
func perpendicular(vec float2) float2 {
	return float2{vec.y, -vec.x}
}

// test if a point p is inside triangle abc
func pointInTriangle(a, b, c, p float2) (inTri bool, weights float3) {
	// test if point is on right side of each segment
	areaABP := signedTriangleArea(a, b, p)
	areaBCP := signedTriangleArea(b, c, p)
	areaCAP := signedTriangleArea(c, a, p)
	inTri = areaABP >= 0 && areaBCP >= 0 && areaCAP >= 0

	// weighting factors (barycentric coordinates)
	invAreaSum := (areaABP + areaBCP + areaCAP)
	weightA := areaBCP / invAreaSum
	weightB := areaCAP / invAreaSum
	weightC := areaABP / invAreaSum
	weights = float3{weightA, weightB, weightC}

	return
}

func randomColor() float3 {
	return float3{rand.Float64(), rand.Float64(), rand.Float64()}
}

type model struct {
	points       []float3
	triangleCols []float3
	transform    transform
}

type transform struct {
	yaw      float64
	pitch    float64
	position float3
}

func (t transform) toWorldPoint(p float3) float3 {
	ihat, jhat, khat := t.getBasisVectors()
	return transformVector(ihat, jhat, khat, p).add(t.position)
}

func (t transform) getBasisVectors() (ihat, jhat, khat float3) {
	// ---- Yaw ----
	ihat_yaw := float3{math.Cos(t.yaw), 0, math.Sin(t.yaw)}
	jhat_yaw := float3{0, 1, 0}
	khat_yaw := float3{-math.Sin(t.yaw), 0, math.Cos(t.yaw)}
	// ---- Pitch ----
	ihat_pitch := float3{1, 0, 0}
	jhat_pitch := float3{0, math.Cos(t.pitch), -math.Sin(t.pitch)}
	khat_pitch := float3{0, math.Sin(t.pitch), math.Cos(t.pitch)}
	// ---- Yaw and Pitch combined ----
	ihat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, ihat_pitch)
	jhat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, jhat_pitch)
	khat = transformVector(ihat_yaw, jhat_yaw, khat_yaw, khat_pitch)

	return
}

func transformVector(ihat, jhat, khat, v float3) float3 {
	return ihat.mulscal(v.x).add(jhat.mulscal(v.y).add(khat.mulscal(v.z)))
}

func signedTriangleArea(a, b, c float2) float64 {
	ac := c.sub(a)
	abPerp := perpendicular(b.sub(a))
	return dot(ac, abPerp) / 2
}
