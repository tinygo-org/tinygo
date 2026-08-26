package main

// Test finalizer invariants that do not require unreachable objects to be collected.
// runtime_asserts checks the finalizer table, count, and bitmap.

import (
	"runtime"
	"sync/atomic"
)

type obj struct{ x int }

const batch = 8

// Finalizers may run concurrently with main under the threads and cores
// schedulers, so all observations shared with a finalizer are atomic.
var (
	clearedRan  atomic.Int32
	replacedRan atomic.Int32
	reachedRan  atomic.Int32
	ranTwice    atomic.Int32
	seen        [batch]atomic.Int32
	reachable   []*obj
)

//go:noinline
func dropCleared() {
	p := &obj{}
	runtime.SetFinalizer(p, func(*obj) { clearedRan.Add(1) })
	runtime.SetFinalizer(p, nil)
}

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

//go:noinline
func keepReachable(id int) {
	p := &obj{x: id}
	runtime.SetFinalizer(p, func(*obj) { reachedRan.Add(1) })
	reachable = append(reachable, p)
}

func main() {
	for i := 0; i < batch; i++ {
		dropCleared()
		dropReplaced(i)
		keepReachable(i)
	}

	// Run GC to check bookkeeping and give finalizers bounded opportunities to run.
	// The checks do not require finalization of an unreachable object.
	for i := 0; i < 8; i++ {
		runtime.GC()
		runtime.Gosched()
	}

	// Keep the reachable objects live across every collection above.
	total := 0
	for _, p := range reachable {
		total += p.x
	}
	if total != batch*(batch-1)/2 {
		println("FAIL: reachable set corrupted:", total)
		return
	}

	switch {
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
