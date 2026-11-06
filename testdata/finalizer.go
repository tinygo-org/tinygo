package main

import (
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

type T struct{ x int }

var (
	ranCount     int
	clearedRan   int
	f1Ran        int
	f2Ran        int
	sink         int
	interiorDone [4]atomic.Int32
)

var finalizerTestMode string

// scrubStack overwrites the stack region used by an alloc-and-drop helper with
// non-pointer words. It must be called at the same call depth as that helper so
// this recursion reuses (and clears) the frame that just held the dropped
// pointer; otherwise a stale copy keeps the object marked and it is never
// collected. The returned value derived from buf keeps the writes live.
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

// allocAndDrop allocates an object, registers a finalizer, and returns without
// leaking any reference to it, so the object becomes unreachable. The finalizer
// must not capture the object (that would pin it forever): it takes the pointer
// as its argument and touches only a package global.
//
//go:noinline
func allocAndDrop() {
	p := &T{x: 42}
	runtime.SetFinalizer(p, func(*T) { ranCount++ })
}

//go:noinline
func allocRegisterClear() {
	p := &T{x: 1}
	runtime.SetFinalizer(p, func(*T) { clearedRan++ })
	runtime.SetFinalizer(p, nil)
}

//go:noinline
func allocRegisterReplace() {
	p := &T{x: 2}
	runtime.SetFinalizer(p, func(*T) { f1Ran++ })
	runtime.SetFinalizer(p, func(*T) { f2Ran++ })
}

// testFires checks that a finalizer runs after its object is collected, and
// only once. scrubStack and the alloc helper are both called here, at the same
// depth, so the scrub clears the helper's stale frame. Gosched lets the
// dedicated finalizer goroutine drain (a no-op under scheduler=none, where
// finalizers already ran inline during GC).
func testFires() {
	allocAndDrop()
	for i := 0; i < 100 && ranCount == 0; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if ranCount == 0 {
		panic("finalizer: never ran after object became unreachable")
	}
	for i := 0; i < 100; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if ranCount != 1 {
		panic("finalizer: ran more than once")
	}
}

// testClear checks that SetFinalizer(obj, nil) removes a finalizer.
func testClear() {
	allocRegisterClear()
	for i := 0; i < 100; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if clearedRan != 0 {
		panic("finalizer: ran after being cleared with nil")
	}
}

// testReplace checks that re-registering replaces the finalizer: only the latest
// one runs, and only once.
func testReplace() {
	allocRegisterReplace()
	for i := 0; i < 100 && f2Ran == 0; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if f1Ran != 0 {
		panic("finalizer: replaced finalizer f1 still ran")
	}
	if f2Ran != 1 {
		panic("finalizer: replacement finalizer f2 did not run exactly once")
	}
}

type interiorObject struct {
	pad   [128]byte
	value int
	data  [8]byte
}

type interiorValue uint16

//go:noinline
func registerInteriorFinalizers() {
	p := &interiorObject{value: 41}
	p.data[1] = 42
	runtime.SetFinalizer(p, func(*interiorObject) { panic("cleared base finalizer ran") })
	runtime.SetFinalizer(&p.value, func(*int) { panic("cleared interior finalizer ran") })
	runtime.SetFinalizer(&p.data[1], func(v *byte) {
		if *v != 42 {
			panic("wrong interior byte finalizer argument")
		}
		interiorDone[2].Add(1)
	})
	runtime.SetFinalizer(&p.data[2], func(*byte) { panic("cleared byte finalizer ran") })
	runtime.SetFinalizer(&p.value, nil)
	runtime.SetFinalizer(&p.data[2], nil)
	runtime.SetFinalizer(&p.data[3], nil)
	runtime.SetFinalizer(p, nil)
	runtime.SetFinalizer(&p.value, func(v *int) {
		if *v != 41 {
			panic("wrong interior int finalizer argument")
		}
		interiorDone[1].Add(1)
	})
	runtime.SetFinalizer(p, func(v *interiorObject) {
		if v.value != 41 {
			panic("wrong base finalizer argument")
		}
		interiorDone[0].Add(1)
	})
	q := new(struct {
		pad   [128]byte
		value interiorValue
	})
	q.value = 43
	runtime.SetFinalizer(&q.value, func(v *interiorValue) {
		if *v != 43 {
			panic("wrong named interior finalizer argument")
		}
		interiorDone[3].Add(1)
	})
}

func testInteriorFinalizers() {
	registerInteriorFinalizers()
	for i := 0; i < 200; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
		if runtime.GOARCH != "wasm" {
			time.Sleep(time.Millisecond)
		}
		done := true
		for j := range interiorDone {
			done = done && interiorDone[j].Load() == 1
		}
		if done {
			return
		}
	}
	panic("interior finalizers did not run exactly once")
}

func main() {
	if finalizerTestMode == "interior" {
		testInteriorFinalizers()
		println("ok")
		return
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "interior":
			testInteriorFinalizers()
			println("ok")
		case "interior-pointer":
			p := new(struct {
				pad    [128]byte
				target *T
			})
			runtime.SetFinalizer(&p.target, func(**T) {})
			runtime.KeepAlive(p)
		case "interior-large":
			p := new(struct {
				pad    [128]byte
				target [16]byte
			})
			runtime.SetFinalizer(&p.target, func(*[16]byte) {})
			runtime.KeepAlive(p)
		}
		return
	}
	testFires()
	testClear()
	testReplace()
	println("ok")
}
