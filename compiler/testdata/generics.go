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

func aliasSize[F float32 | float64]() uintptr {
	return unsafe.Sizeof(F(0))
}

func aliasSize32() uintptr {
	type F = float32
	return aliasSize[F]()
}

func aliasSize64() uintptr {
	type F = float64
	return aliasSize[F]()
}

type aliasMethodResult[F float32 | float64] struct {
	Value F
}

type aliasMethodValue[F float32 | float64] struct{}

func (aliasMethodValue[F]) Get() aliasMethodResult[F] {
	return aliasMethodResult[F]{}
}

func main() {
	var af, bf Point[float32]
	Add(af, bf)

	var ai, bi Point[int]
	Add(ai, bi)

	checkSize(aliasSize32())
	checkSize(aliasSize64())
}

func checkSize(uintptr)

func checkBool(bool)

func aliasMethod32(x any) {
	type F = float32
	_, ok := x.(interface {
		Get() aliasMethodResult[F]
	})
	checkBool(ok)
}
