package main

// Test finalizer registration, replacement, removal, and reused heap addresses.
// The wasm target provides deterministic finalization for these tests.

import "runtime"

type box struct{ x int }

const batch = 32

var (
	reregisteredRan int
	replacedOldRan  int
	replacedNewRan  int
	churnRan        int
	reuseFirstRan   int
	reuseSecondRan  int
	keptRan         int
	droppedRan      int
	sink            int
)

// scrubStack removes stale pointers from the helper frame so collection is deterministic.
// Call it at the same call depth as the allocation helper.
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
func allocClearThenRegister() {
	p := &box{x: 1}
	runtime.SetFinalizer(p, func(*box) { panic("cleared finalizer ran") })
	runtime.SetFinalizer(p, nil)
	runtime.SetFinalizer(p, func(*box) { reregisteredRan++ })
}

func testClearThenRegister() {
	allocClearThenRegister()
	for i := 0; i < 200 && reregisteredRan == 0; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if reregisteredRan != 1 {
		panic("finalizerbits: re-registered finalizer did not run exactly once")
	}
}

//go:noinline
func allocRegisterTwice() {
	for i := 0; i < batch; i++ {
		p := &box{x: i}
		runtime.SetFinalizer(p, func(*box) { replacedOldRan++ })
		runtime.SetFinalizer(p, func(*box) { replacedNewRan++ })
	}
}

func testRegisterTwiceLeavesOne() {
	allocRegisterTwice()
	for i := 0; i < 200 && replacedNewRan < batch; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if replacedOldRan != 0 {
		panic("finalizerbits: replaced finalizer still ran")
	}
	if replacedNewRan != batch {
		panic("finalizerbits: replacement did not run exactly once per object")
	}
}

//go:noinline
func allocChurn() {
	p := &box{x: 3}
	for i := 0; i < 64; i++ {
		runtime.SetFinalizer(p, func(*box) { churnRan++ })
		runtime.SetFinalizer(p, nil)
	}
}

func testChurnLeavesNothing() {
	allocChurn()
	for i := 0; i < 200; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if churnRan != 0 {
		panic("finalizerbits: churned register/clear left a live registration")
	}
}

//go:noinline
func allocFirstRound() {
	for i := 0; i < batch; i++ {
		p := &box{x: i}
		runtime.SetFinalizer(p, func(*box) { reuseFirstRan++ })
	}
}

//go:noinline
func allocSecondRound() {
	for i := 0; i < batch; i++ {
		p := &box{x: i}
		runtime.SetFinalizer(p, func(*box) { reuseSecondRan++ })
	}
}

func testAddressReuse() {
	allocFirstRound()
	for i := 0; i < 200 && reuseFirstRan < batch; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if reuseFirstRan != batch {
		panic("finalizerbits: first round did not run every finalizer")
	}
	allocSecondRound()
	for i := 0; i < 200 && reuseSecondRan < batch; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if reuseSecondRan != batch {
		panic("finalizerbits: second round into reused memory lost finalizers")
	}
}

//go:noinline
func allocMixedBatch() {
	for i := 0; i < batch; i++ {
		p := &box{x: i}
		if i%2 == 0 {
			runtime.SetFinalizer(p, func(*box) { droppedRan++ })
			runtime.SetFinalizer(p, nil)
		} else {
			runtime.SetFinalizer(p, func(*box) { keptRan++ })
		}
	}
}

func testMixedBatch() {
	allocMixedBatch()
	for i := 0; i < 200 && keptRan < batch/2; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if droppedRan != 0 {
		panic("finalizerbits: a cleared finalizer inside the batch ran")
	}
	if keptRan != batch/2 {
		panic("finalizerbits: kept finalizers did not all run exactly once")
	}
}

func main() {
	testClearThenRegister()
	testRegisterTwiceLeavesOne()
	testChurnLeavesNothing()
	testAddressReuse()
	testMixedBatch()
	println("ok")
}
