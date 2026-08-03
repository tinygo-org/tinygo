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
	"time"
)

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

// finalizerRuns is the total number of finalizer invocations observed so far,
// across every counter. Individual counters are asserted on at the end; this
// sum exists only to tell "the runner is still working" from "the queue is
// empty".
func finalizerRuns() int {
	n := clearedRan + replacedRan + reachedRan + ranTwice
	for _, s := range seen {
		n += s
	}
	return n
}

// Bounds for drainFinalizers. quietRounds is how many consecutive rounds must
// observe no new invocation before the queue counts as drained; maxRounds caps
// a target that never runs a finalizer at all, which the vacuity check in main
// then reports.
//
// Measured, every target here drains in 4 rounds and then 3, including
// scheduler.threads, so maxRounds is headroom for a loaded machine rather than
// an expected cost: the loop exits on quiescence long before reaching it.
const (
	quietRounds = 3
	maxRounds   = 50
)

// drainFinalizers collects until the finalizer queue is drained, and returns
// only once it is.
//
// It waits on the observable result rather than on a fixed delay. Gosched does
// not synchronize with the runner under scheduler.threads, where it is a no-op
// (every goroutine is its own thread, so there is nothing to yield to) and the
// runner is a separate thread blocked on a futex. Sleeping a fixed amount would
// only make the race less likely; polling until invocations stop arriving is
// what actually establishes that the queue is empty. The sleep below is the
// poll interval, not the wait.
//
// Quiescence alone is not enough to start with, because a runner that has not
// been scheduled yet looks identical to a drained queue. So the quiet rounds
// only count once at least one finalizer has run.
func drainFinalizers() {
	quiet := 0
	for i := 0; i < maxRounds; i++ {
		before := finalizerRuns()
		runtime.GC()
		runtime.Gosched()
		time.Sleep(time.Millisecond)
		switch {
		case finalizerRuns() != before:
			quiet = 0
		case before == 0:
			// Nothing has run yet: cannot tell a drained queue from a runner
			// that has not started, so keep polling.
		default:
			quiet++
			if quiet == quietRounds {
				return
			}
		}
	}
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

	drainFinalizers()
	scrubStack(12)
	drainFinalizers()

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

	// Count the replacement finalizers that ran. The assertions below are all of
	// the "must never run" kind, so they are only meaningful if something was
	// collected and drained at all: with nothing collected they hold trivially
	// and the test reports success while checking nothing. Requiring at least
	// one firing turns that silent pass into a failure. It stays at "at least
	// one" rather than "all", because a conservative stack scan is allowed to
	// pin any individual object.
	replacementsRan := 0
	for _, n := range seen {
		replacementsRan += n
	}

	switch {
	case replacementsRan == 0:
		println("FAIL: no finalizer ran at all, the assertions below prove nothing")
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
