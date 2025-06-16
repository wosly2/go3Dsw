package raster

import (
	"math"
	"math/rand"
)

type Float3 struct {
	X, Y, Z float64
}

type Float2 struct {
	X, Y float64
}

func (v Float2) add(other Float2) Float2 {
	return Float2{v.X + other.X, v.Y + other.Y}
}

func (v Float2) addscal(scalar float64) Float2 {
	return Float2{v.X + scalar, v.Y + scalar}
}

func (v Float2) mul(other Float2) Float2 {
	return Float2{v.X * other.X, v.Y * other.Y}
}

func (v Float2) sub(other Float2) Float2 {
	return Float2{v.X - other.X, v.Y - other.Y}
}

func (v Float2) mulscal(scalar float64) Float2 {
	return Float2{v.X * scalar, v.Y * scalar}
}

func (v Float3) add(other Float3) Float3 {
	return Float3{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Float3) sub(other Float3) Float3 {
	return Float3{v.X - other.X, v.Y - other.Y, v.Z - other.Z}
}

func (v Float3) mulscal(scalar float64) Float3 {
	return Float3{v.X * scalar, v.Y * scalar, v.Z * scalar}
}

func (v Float3) mul(other Float3) Float3 {
	return Float3{v.X * other.X, v.Y * other.Y, v.Z * other.Z}
}

func (v Float3) under(scalar float64) Float3 {
	return Float3{scalar / v.X, scalar / v.Y, scalar / v.Z}
}

func (v Float3) make2() Float2 {
	return Float2{v.X, v.Y}
}

func (v Float3) magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Float3) normalized() Float3 {
	mag := v.magnitude()
	if mag == 0 {
		return Float3{0, 0, 0}
	}
	return v.mulscal(1 / mag)
}

func lerp(a, b Float3, p float64) Float3 {
	return Float3{
		a.X*(1-p) + b.X*p,
		a.Y*(1-p) + b.Y*p,
		a.Z*(1-p) + b.Z*p,
	}
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
	r := uint32(clamp(f.X, 0, 1) * 255)
	g := uint32(clamp(f.Y, 0, 1) * 255)
	b := uint32(clamp(f.Z, 0, 1) * 255)
	a := uint32(255) // set default alpha

	return (a << 24) | (r << 16) | (g << 8) | b
}

// dot product of a and b
// equal to product of lengths * cosine of their angle
func dot(a, b Float2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func cross3(a, b Float3) Float3 {
	return Float3{
		(a.Y * b.Z) - (a.Z * b.Y),
		(a.Z * b.X) - (a.X * b.Z),
		(a.X * b.Y) - (a.Y * b.X),
	}
}

// dot product of a and b, but float3!
func dot3(a, b Float3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// finds perpendicular vector (90 clockwise)
func perpendicular(vec Float2) Float2 {
	return Float2{vec.Y, -vec.X}
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

func RandomColor() Float3 {
	return Float3{rand.Float64(), rand.Float64(), rand.Float64()}
}

func transformVector(ihat, jhat, khat, v Float3) Float3 {
	return ihat.mulscal(v.X).add(jhat.mulscal(v.Y).add(khat.mulscal(v.Z)))
}

func signedTriangleArea(a, b, c Float2) float64 {
	ac := c.sub(a)
	abPerp := perpendicular(b.sub(a))
	return dot(ac, abPerp) / 2
}

func ToRadians(d float64) float64 {
	return d * math.Pi / 180
}
