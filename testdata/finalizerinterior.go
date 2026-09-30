package main

import "runtime"

type interiorOuter struct {
	head  [64]byte
	inner [64]byte
}

const interiorObjects = 64

var (
	interiorRan  int
	interiorGot  int
	interiorSink int
)

//go:noinline
func scrubInteriorStack(depth int) int {
	if depth == 0 {
		return interiorSink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	interiorSink += buf[depth&63]
	return scrubInteriorStack(depth-1) + buf[0]
}

// The finalizer is set on a field inside the allocation, not on its start.
//
//go:noinline
func registerInteriorFinalizers() {
	for i := 0; i < interiorObjects; i++ {
		outer := &interiorOuter{}
		outer.inner[0] = 0x5A
		runtime.SetFinalizer(&outer.inner, func(inner *[64]byte) {
			interiorRan++
			if inner[0] == 0x5A {
				interiorGot++
			}
		})
	}
}

func main() {
	registerInteriorFinalizers()
	for i := 0; i < 100 && interiorRan < interiorObjects; i++ {
		interiorSink += scrubInteriorStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	// A stale stack pointer can keep one object alive.
	if interiorRan < interiorObjects-1 {
		println("interior finalizers ran", interiorRan, "of", interiorObjects)
		panic("interior pointer finalizers did not run")
	}
	if interiorGot != interiorRan {
		panic("finalizer did not get the registered pointer")
	}
	println("ok")
}
