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

import (
	"runtime"
	"sync/atomic"
	"time"
)

type obj struct{ x int }

const batch = 8

// The counters are written by finalizers and read by main. Under
// scheduler.threads and scheduler.cores those are different threads running at
// the same time, so every access goes through sync/atomic rather than a plain
// int.
var (
	clearedRan  atomic.Int32
	replacedRan atomic.Int32
	reachedRan  atomic.Int32
	ranTwice    atomic.Int32
	seen        [batch]atomic.Int32
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
	runtime.SetFinalizer(p, func(*obj) { clearedRan.Add(1) })
	runtime.SetFinalizer(p, nil)
}

// dropReplaced registers a finalizer and then replaces it. Registering twice
// must replace rather than accumulate, so the first func may never run and the
// second may run at most once.
//
//go:noinline
func dropReplaced(id int) {
	p := &obj{}
	runtime.SetFinalizer(p, func(*obj) { replacedRan.Add(1) })
	runtime.SetFinalizer(p, func(*obj) {
		if seen[id].Add(1) > 1 {
			ranTwice.Add(1)
		}
	})
}

// keepReachable registers a finalizer on an object held by a global. A
// reachable object must never be finalized.
//
//go:noinline
func keepReachable(id int) {
	p := &obj{x: id}
	runtime.SetFinalizer(p, func(*obj) { reachedRan.Add(1) })
	reachable = append(reachable, p)
}

// replacementsRan is how many of the batch replacement finalizers have run.
func replacementsRan() int {
	n := 0
	for i := range seen {
		n += int(seen[i].Load())
	}
	return n
}

// maxRounds bounds waitForDrain. Measured, every target here reaches the full
// count within a handful of rounds, so this is headroom for a loaded machine
// rather than an expected cost.
const maxRounds = 500

// waitForDrain collects until every replacement finalizer has run, and reports
// whether it got there.
//
// It waits for a known count rather than for the queue to look idle. Idleness
// cannot be observed from here: runtime exposes no way to ask whether the
// finalizer queue is empty, and under scheduler.threads the runner is a
// separate thread, so a stretch with no new invocation is indistinguishable
// from a runner that has simply not been scheduled yet. Waiting for a specific
// number of invocations has no such ambiguity, and not reaching it is a test
// failure rather than a silently short wait.
//
// batch is the right target because these objects are allocated and dropped
// inside a //go:noinline helper whose frame scrubStack then overwrites, so
// nothing is left pointing at them for a conservative scan to find.
func waitForDrain() bool {
	for i := 0; i < maxRounds; i++ {
		runtime.GC()
		runtime.Gosched()
		time.Sleep(time.Millisecond)
		if replacementsRan() == batch {
			// The runner has worked through the queue these objects were in. Do
			// one more pass so that a finalizer which must NOT run, but which a
			// wrong bitmap bit left registered, is queued and drained here
			// instead of after the counters are read.
			runtime.GC()
			runtime.Gosched()
			time.Sleep(time.Millisecond)
			return true
		}
	}
	return false
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

	drained := waitForDrain()

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
	case !drained:
		println("FAIL: only", replacementsRan(), "of", batch, "replacement finalizers ran, the assertions below prove nothing")
	case clearedRan.Load() != 0:
		println("FAIL: cleared finalizer ran:", clearedRan.Load())
	case replacedRan.Load() != 0:
		println("FAIL: replaced finalizer ran:", replacedRan.Load())
	case reachedRan.Load() != 0:
		println("FAIL: reachable object was finalized:", reachedRan.Load())
	case ranTwice.Load() != 0:
		println("FAIL: finalizer ran more than once:", ranTwice.Load())
	default:
		println("ok")
	}
}
