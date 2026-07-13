package main

// Tests for runtime.SetFinalizer on the block GC.
//
// This test is only run on the precise wasm/wasi targets (see the tests slice
// and the skip in main_test.go): they track stack pointers precisely and
// reliably collect a dropped object, so the finalizer deterministically fires.
// The host default GC is boehm, where SetFinalizer is a no-op, and conservative
// stack scanning on the emulated targets cannot reliably collect the object, so
// firing cannot be asserted there.

import "runtime"

type T struct{ x int }

var (
	ranCount   int
	clearedRan int
	f1Ran      int
	f2Ran      int
	sink       int
)

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

func main() {
	testFires()
	testClear()
	testReplace()
	println("ok")
}
