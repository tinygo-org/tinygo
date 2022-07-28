package main

type Coord interface {
	int | float32
}

type Point[T Coord] struct {
	X, Y T
}

func Add[T Coord](a, b Point[T]) Point[T] {
	return Point[T]{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func main() {
	var af, bf Point[float32]
	Add(af, bf)
}
