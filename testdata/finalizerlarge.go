package main

import "runtime"

type largeFinalizerObject struct {
	data [128]byte
}

var largeFinalizerRan bool
var largeFinalizerSink int

//go:noinline
func scrubLargeFinalizerStack(depth int) int {
	if depth == 0 {
		return largeFinalizerSink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	return scrubLargeFinalizerStack(depth-1) + buf[0]
}

//go:noinline
func registerLargeFinalizer() {
	p := new(largeFinalizerObject)
	runtime.SetFinalizer(p, func(*largeFinalizerObject) {
		largeFinalizerRan = true
	})
}

func main() {
	registerLargeFinalizer()
	for i := 0; i < 100 && !largeFinalizerRan; i++ {
		largeFinalizerSink += scrubLargeFinalizerStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if !largeFinalizerRan {
		panic("large object finalizer did not run")
	}
	println("ok")
}
