package main

// Invariants of runtime.SetFinalizer that hold on every target the block GC
// supports, not only the ones where a dropped object is deterministically
// collected.
//
// finalizer.go, finalizerbits.go and finalizeridle.go all assert that a
// finalizer fired, which needs the dropped object to actually be collected, so
// they only run on wasm (see the skip in main_test.go). Conservative stack
// scanning elsewhere can keep a dropped object alive and the finalizer then
// correctly does not run.
//
// The opposite direction is portable: a conservative collector only ever
// over-retains, never under-retains, so "this finalizer must never run" holds
// on every target. Those are exactly the invariants the per-block registration
// bitmap can break, because a wrong bit skips the table walk that the clear and
// replace semantics of SetFinalizer depend on. A stale or torn bit therefore
// shows up here as a finalizer that runs when it must not.

import "runtime"

type obj struct{ x int }

const batch = 8

var (
	clearedRan  int
	replacedRan int
	reachedRan  int
	ranTwice    int
	seen        [batch]int
	reachable   []*obj
	sink        int
)

// scrubStack overwrites the stack region used by an alloc-and-drop helper with
// non-pointer words, so a stale frame does not keep the dropped object marked.
// It must be called at the same call depth as those helpers.
//
//go:noinline
func scrubStack(depth int) int {
	if depth <= 0 {
		return sink
	}
	var buf [16]int
	for i := range buf {
		buf[i] = depth + i
	}
	sink += buf[depth&15]
	return scrubStack(depth-1) + buf[0]
}

// dropCleared registers a finalizer, clears it, then drops the object. Clearing
// must remove the registration, so this finalizer may never run.
//
//go:noinline
func dropCleared() {
	p := &obj{}
	runtime.SetFinalizer(p, func(*obj) { clearedRan++ })
	runtime.SetFinalizer(p, nil)
}

// dropReplaced registers a finalizer and then replaces it. Registering twice
// must replace rather than accumulate, so the first func may never run and the
// second may run at most once.
//
//go:noinline
func dropReplaced(id int) {
	p := &obj{}
	runtime.SetFinalizer(p, func(*obj) { replacedRan++ })
	runtime.SetFinalizer(p, func(*obj) {
		seen[id]++
		if seen[id] > 1 {
			ranTwice++
		}
	})
}

// keepReachable registers a finalizer on an object held by a global. A
// reachable object must never be finalized.
//
//go:noinline
func keepReachable(id int) {
	p := &obj{x: id}
	runtime.SetFinalizer(p, func(*obj) { reachedRan++ })
	reachable = append(reachable, p)
}

func main() {
	for i := 0; i < batch; i++ {
		dropCleared()
	}
	scrubStack(12)

	for i := 0; i < batch; i++ {
		dropReplaced(i)
	}
	scrubStack(12)

	for i := 0; i < batch; i++ {
		keepReachable(i)
	}

	// Collect repeatedly, yielding so the finalizer runner goroutine gets to
	// drain anything that was queued.
	for i := 0; i < 4; i++ {
		runtime.GC()
		runtime.Gosched()
	}
	scrubStack(12)
	for i := 0; i < 4; i++ {
		runtime.GC()
		runtime.Gosched()
	}

	// Touch the reachable set after the collections so it stays a live root
	// across all of them.
	total := 0
	for _, p := range reachable {
		total += p.x
	}
	if total != batch*(batch-1)/2 {
		println("FAIL: reachable set corrupted:", total)
		return
	}

	switch {
	case clearedRan != 0:
		println("FAIL: cleared finalizer ran:", clearedRan)
	case replacedRan != 0:
		println("FAIL: replaced finalizer ran:", replacedRan)
	case reachedRan != 0:
		println("FAIL: reachable object was finalized:", reachedRan)
	case ranTwice != 0:
		println("FAIL: finalizer ran more than once:", ranTwice)
	default:
		println("ok")
	}
}
