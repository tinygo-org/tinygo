package main

// Test idle finalizer collection and the lifetime of blocked and completed asyncify stacks.
// The wasm target provides deterministic finalization for these tests.

import (
	"runtime"
	"sync/atomic"
	"time"
)

// batch must exceed the finalizer registration threshold to trigger idle collection.
const batch = 64

var (
	ranDropped int
	ranOnStack int
	ranInArgs  int
	sink       int

	blockedRan [3]atomic.Int32
	controlRan atomic.Int32
)

type blockedObject struct{ x int }

//go:noinline
func blockOperation(kind int, ready chan<- struct{}, ch chan struct{}) {
	ready <- struct{}{}
	switch kind {
	case 0:
		select {}
	case 1:
		ch <- struct{}{}
	case 2:
		<-ch
	}
}

//go:noinline
func holdWhileBlocked(kind int, ready chan<- struct{}, ch chan struct{}) {
	p := &blockedObject{x: kind}
	runtime.SetFinalizer(p, func(*blockedObject) { blockedRan[kind].Add(1) })
	blockOperation(kind, ready, ch)
	// blockOperation can return, so p remains live on this suspended stack.
	runtime.KeepAlive(p)
}

//go:noinline
func dropProgressControl() {
	p := &blockedObject{x: 8}
	runtime.SetFinalizer(p, func(*blockedObject) { controlRan.Add(1) })
}

// testPermanentlyBlockedStacks checks that blocked task stacks remain GC roots.
// A control finalizer confirms that GC and finalizer processing made progress.
func testPermanentlyBlockedStacks() {
	ready := make(chan struct{}, 3)
	go holdWhileBlocked(0, ready, nil) // select{}
	go holdWhileBlocked(1, ready, nil) // nil channel send
	go holdWhileBlocked(2, ready, nil) // nil channel receive
	<-ready
	<-ready
	<-ready

	dropProgressControl()
	for i := 0; i < 100 && controlRan.Load() == 0; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
	}
	if controlRan.Load() != 1 {
		panic("control finalizer did not prove GC progress")
	}
	// Collect once more so temporary scheduler roots cannot hide an unrooted
	// blocked task during the control collection.
	runtime.GC()
	runtime.Gosched()
	for i, name := range [...]string{"select{}", "nil-channel send", "nil-channel receive"} {
		if blockedRan[i].Load() != 0 {
			panic(name + " stack-held object was finalized")
		}
	}
}

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

// registerAndDrop creates unreachable objects with finalizers that do not capture them.
// This allows the idle GC to collect the objects.
//
//go:noinline
func registerAndDrop() {
	for i := 0; i < batch; i++ {
		p := new([2]int)
		runtime.SetFinalizer(p, func(*[2]int) { ranDropped++ })
	}
}

func testIdleCollect() {
	registerAndDrop()
	for i := 0; i < 500 && ranDropped < batch; i++ {
		sink += scrubStack(40)
		time.Sleep(time.Millisecond)
	}
	if ranDropped != batch {
		panic("idle collection did not run every finalizer")
	}
}

func testFinishedGoroutineStacks() {
	done := make(chan struct{})
	for i := 0; i < batch; i++ {
		go func() {
			p := new([2]int)
			runtime.SetFinalizer(p, func(*[2]int) { ranOnStack++ })
			// p stays on this goroutine's stack until it returns just below.
			done <- struct{}{}
		}()
	}
	for i := 0; i < batch; i++ {
		<-done
	}
	for i := 0; i < 500 && ranOnStack < batch; i++ {
		sink += scrubStack(40)
		time.Sleep(time.Millisecond)
	}
	if ranOnStack != batch {
		panic("finished goroutine stack still pinned finalized objects")
	}
}

// launchArgGoroutine passes an object through the task argument bundle.
// The caller returns so scrubStack can remove its transient pointer.
//
//go:noinline
func launchArgGoroutine(done chan struct{}) {
	p := new([2]int)
	runtime.SetFinalizer(p, func(*[2]int) { ranInArgs++ })
	go func(q *[2]int) {
		sink += q[0]
		done <- struct{}{}
	}(p)
}

func testFinishedGoroutineArgs() {
	done := make(chan struct{})
	for i := 0; i < batch; i++ {
		launchArgGoroutine(done)
	}
	for i := 0; i < batch; i++ {
		<-done
	}
	for i := 0; i < 500 && ranInArgs < batch; i++ {
		sink += scrubStack(40)
		time.Sleep(time.Millisecond)
	}
	if ranInArgs != batch {
		panic("finished goroutine args still pinned finalized objects")
	}
}

func main() {
	testPermanentlyBlockedStacks()
	testIdleCollect()
	testFinishedGoroutineStacks()
	testFinishedGoroutineArgs()
	println("ok")
}
