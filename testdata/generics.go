package main

import (
	"github.com/tinygo-org/tinygo/testdata/generics/testa"
	"github.com/tinygo-org/tinygo/testdata/generics/testb"
)

func main() {
	println("add:", Add(3, 5))
	println("add:", Add(int8(3), 5))

	var c C[int]
	c.F() // issue 2951

	testa.Test()
	testb.Test()
}

type Integer interface {
	int | int8 | int16 | int32 | int64
}

func Add[T Integer](a, b T) T {
	return a + b
}

// Test for https://github.com/tinygo-org/tinygo/issues/2951
type C[V any] struct{}

func (c *C[V]) F() {}
