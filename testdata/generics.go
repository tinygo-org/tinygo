package main

func main() {
	println("add:", Add(3, 5))
	println("add:", Add(int8(3), 5))
}

type Integer interface {
	int | int8 | int16 | int32 | int64
}

func Add[T Integer](a, b T) T {
	return a + b
}
