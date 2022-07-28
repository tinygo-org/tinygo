package main

import "unsafe"

type Coord interface {
	int | float32
}

type Point[T Coord] struct {
	X, Y T
}

func Add[T Coord](a, b Point[T]) Point[T] {
	checkSize(unsafe.Alignof(a))
	checkSize(unsafe.Sizeof(a))
	return Point[T]{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func main() {
	var af, bf Point[float32]
	Add(af, bf)

	var ai, bi Point[int]
	Add(ai, bi)
}

func checkSize(uintptr)
