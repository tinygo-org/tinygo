package main

type color struct {
	r, g, b uint8
}

// A simple aggregate comparison should be performed directly.
func colorEq(x, y color) bool {
	return x == y
}

type sprite [16][16]color

// A huge binary aggregate with no padding should be compared with runtime.memEqualBytes.
func spriteEq(x, y *sprite) bool {
	return *x == *y
}

// An interface array comparison should always perform all comparisons.
func interfaceArrayEq(x, y *[5]any) bool {
	return *x == *y
}

type point3D struct {
	x, y, z float32
}

type triangle [3]point3D

// A large aggregate of floats should be compared with a loop.
func triangleEq(x, y *triangle) bool {
	return *x == *y
}

type options struct {
	a bool
	b uint64
	c string
	d any
	e float32
	f complex128
	g error
}

// A large struct of mixed types should be compared in memory with a helper function.
func optionsEq(x, y *options) bool {
	return *x == *y
}

type misc struct {
	f32 float32
	f64 float64
}

type moreOptions struct {
	misc
	options
	h string
}

// A large nested struct should reuse the comparison functions of the inner struct(s).
func moreOptionsEq(x, y *moreOptions) bool {
	return *x == *y
}

type padded struct {
	a [5]any
	_ [16]byte
}

// Blank ident fields should be excluded from comparison.
func paddedEq(x, y *padded) bool {
	return *x == *y
}

// Maps should use the same comparison logic and tricks for hashing.
var _ = map[color]struct{}{}
var _ = map[sprite]struct{}{}
var _ = map[point3D]struct{}{}
var _ = map[triangle]struct{}{}
var _ = map[moreOptions]struct{}{}
var _ = map[padded]struct{}{}
