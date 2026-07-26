package main

// Tests the registration bookkeeping behind runtime.SetFinalizer on the block
// GC: the per-block bit that records whether an object already has a finalizer.
// The bit is what lets a fresh object skip the registered-finalizer scan, so
// these cases pin the invariants that skipping must never break:
//
//   - an object whose finalizer was cleared and then registered again still runs
//     it exactly once, so clearing resets the bookkeeping;
//   - registering twice replaces, it never leaves two registrations behind
//     (which would run the finalizer twice);
//   - churning register/clear on one object leaves no residue;
//   - memory reused by a later object registers correctly, so a dead object's
//     bookkeeping does not leak onto whatever lands at its address next;
//   - a batch where only some objects keep a finalizer runs exactly those.
//
// Like finalizer.go, this is only run on the precise wasm target (see the tests
// slice and the skip in main_test.go): there a dropped object is deterministically
// collected, so the finalizers fire predictably.
//
// Each test calls its alloc helper and scrubStack at the same call depth, so the
// recursion reuses and clears the frame that just held the dropped pointers.

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

//go:noinline
func allocClearThenRegister() {
	p := &box{x: 1}
	runtime.SetFinalizer(p, func(*box) { panic("cleared finalizer ran") })
	runtime.SetFinalizer(p, nil)
	runtime.SetFinalizer(p, func(*box) { reregisteredRan++ })
}

// testClearThenRegister checks that clearing a finalizer and registering a new
// one leaves exactly the new one: clearing has to reset the bookkeeping, not
// just unlink the entry.
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

// testRegisterTwiceLeavesOne checks the replace path over a whole batch: the
// second registration must find the first one and take its place. A missed
// lookup would leave two registrations for the same object, and its finalizer
// would run twice.
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

// testChurnLeavesNothing checks that many register/clear rounds on one object
// leave nothing behind: the object dies with no finalizer, so nothing runs.
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

// testAddressReuse checks that objects allocated into memory freed by a previous
// finalized batch register correctly themselves. A dead object's bookkeeping must
// not survive onto whatever lands at its address next.
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

// testMixedBatch checks that clearing some registrations inside a batch affects
// only those objects: the ones still registered run, the cleared ones do not.
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
