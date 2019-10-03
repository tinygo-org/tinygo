package main

import "unsafe"

var global = 3

//go:linkname foo foo
func foo(ptr unsafe.Pointer) {
	if ptr == unsafe.Pointer(&global) {
		println("pointer matches")
	} else {
		println("pointer mismatch!")
	}
}

//go:linkname foobar foo
func foobar(ptr *int)

//go:linkname testMultiReturn multiret
func testMultiReturn(x int) (*int, *int) {
	for i := 2; i < x; i++ {
		if x%i == 0 {
			x /= i
			return &x, &i
		}
	}
	one := 1
	return &x, &one
}

//go:linkname testMultiReturnCast multiret
func testMultiReturnCast(int) (unsafe.Pointer, unsafe.Pointer)

func main() {
	foo(nil)
	foobar(&global)
	a, b := testMultiReturn(7)
	println(*a, *b)
	x, y := testMultiReturnCast(6)
	println(*(*int)(x), *(*int)(y))
}
