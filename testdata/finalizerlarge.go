package main

import "runtime"

type largeFinalizerObject struct {
	data [128]byte
}

var (
	largeFinalizerRan bool
	sink              int
)

// scrubStack removes stale pointers from the helper frame so collection is deterministic.
//
//go:noinline
func scrubStack(depth int) int {
	if depth <= 0 {
		return sink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	sink += buf[depth&63]
	return scrubStack(depth-1) + buf[0]
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
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if !largeFinalizerRan {
		panic("large object finalizer did not run")
	}
	println("ok")
}
