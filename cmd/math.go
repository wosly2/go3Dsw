package main

import (
	"math"
	"math/rand"
)

type Float3 struct {
	x, y, z float64
}

type Float2 struct {
	x, y float64
}

func (v Float2) add(other Float2) Float2 {
	return Float2{v.x + other.x, v.y + other.y}
}

func (v Float2) sub(other Float2) Float2 {
	return Float2{v.x - other.x, v.y - other.y}
}

func (v Float2) mulscal(scalar float64) Float2 {
	return Float2{v.x * scalar, v.y * scalar}
}

func (v Float3) add(other Float3) Float3 {
	return Float3{v.x + other.x, v.y + other.y, v.z + other.z}
}

func (v Float3) sub(other Float3) Float3 {
	return Float3{v.x - other.x, v.y - other.y, v.z - other.z}
}

func (v Float3) mulscal(scalar float64) Float3 {
	return Float3{v.x * scalar, v.y * scalar, v.z * scalar}
}

func (v Float3) under(scalar float64) Float3 {
	return Float3{scalar / v.x, scalar / v.y, scalar / v.z}
}

func (v Float3) make2() Float2 {
	return Float2{v.x, v.y}
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
func (f Float3) toUint32() (u uint32) {
	// ensure they're clipped to 0-1
	r := uint32(clamp(f.x, 0, 1) * 255)
	g := uint32(clamp(f.y, 0, 1) * 255)
	b := uint32(clamp(f.z, 0, 1) * 255)
	a := uint32(255) // set default alpha

	return (a << 24) | (r << 16) | (g << 8) | b
}

// dot product of a and b
// equal to product of lengths * cosine of their angle
func dot(a, b Float2) float64 {
	return a.x*b.x + a.y*b.y
}

// dot product of a and b, but float3!
func dot3(a, b Float3) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

// finds perpendicular vector (90 clockwise)
func perpendicular(vec Float2) Float2 {
	return Float2{vec.y, -vec.x}
}

// test if a point p is inside triangle abc
func pointInTriangle(a, b, c, p Float2) (inTri bool, weights Float3) {
	// test if point is on right side of each segment
	areaABP := signedTriangleArea(a, b, p)
	areaBCP := signedTriangleArea(b, c, p)
	areaCAP := signedTriangleArea(c, a, p)
	inTri = areaABP >= 0 && areaBCP >= 0 && areaCAP >= 0

	// weighting factors (barycentric coordinates)
	totalArea := (areaABP + areaBCP + areaCAP)
	weightA := areaBCP / totalArea
	weightB := areaCAP / totalArea
	weightC := areaABP / totalArea
	weights = Float3{weightA, weightB, weightC}

	inTri = inTri && totalArea > 0

	return
}

func randomColor() Float3 {
	return Float3{rand.Float64(), rand.Float64(), rand.Float64()}
}

func transformVector(ihat, jhat, khat, v Float3) Float3 {
	return ihat.mulscal(v.x).add(jhat.mulscal(v.y).add(khat.mulscal(v.z)))
}

func signedTriangleArea(a, b, c Float2) float64 {
	ac := c.sub(a)
	abPerp := perpendicular(b.sub(a))
	return dot(ac, abPerp) / 2
}

func toRadians(d float64) float64 {
	return d * math.Pi / 180
}
