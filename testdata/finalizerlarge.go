package main

import "runtime"

type largeFinalizerObject struct {
	data [128]byte
}

var largeFinalizerRan bool

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
		runtime.GC()
		runtime.Gosched()
	}
	if !largeFinalizerRan {
		panic("large object finalizer did not run")
	}
	println("ok")
}
