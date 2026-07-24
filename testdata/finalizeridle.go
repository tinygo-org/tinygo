package main

// Tests that the cooperative scheduler reclaims finalizer-guarded objects on its
// own, without an explicit runtime.GC(), once enough finalizers have been
// registered since the last collection. A registered finalizer usually guards an
// external resource whose Go-heap cost is tiny relative to what it pins, so the
// registration count drives a proactive collection at the scheduler's idle
// point. The second case additionally checks that a finished goroutine's stack
// no longer pins the objects its frames held.
//
// Like finalizer.go, this is only run on the precise wasm target (see the tests
// slice and the skip in main_test.go): there a dropped object is deterministically
// collected, so the finalizers fire predictably. It never calls runtime.GC(): the
// point is that the idle-point trigger collects on its own.

import (
	"runtime"
	"time"
)

// batch must exceed the runtime's finalizer-registration threshold so the idle
// collection is guaranteed to trigger.
const batch = 64

var (
	ranDropped int
	ranOnStack int
	ranInArgs  int
	sink       int
)

// scrubStack overwrites the stack region used by an alloc-and-drop helper with
// non-pointer words. It is called at the same depth as that helper so this
// recursion reuses (and clears) the frame that just held the dropped pointers;
// otherwise a stale copy keeps an object marked and it is never collected. The
// returned value derived from buf keeps the writes live.
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

// registerAndDrop registers `batch` finalizers and returns without leaking any
// reference to the objects, so they become unreachable. The finalizer must not
// capture its object (that would pin it forever): it takes the pointer as its
// argument and touches only a package global.
//
//go:noinline
func registerAndDrop() {
	for i := 0; i < batch; i++ {
		p := new([2]int)
		runtime.SetFinalizer(p, func(*[2]int) { ranDropped++ })
	}
}

// testIdleCollect checks that registering many finalizers and then only parking
// the goroutine (time.Sleep, never runtime.GC()) is enough for the objects to be
// collected and their finalizers to run.
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

// testFinishedGoroutineStacks checks that a goroutine which registers a finalizer
// on a stack-local object and then returns no longer pins that object: once the
// goroutine has finished, the idle collection reclaims the object. Without
// zeroing a finished goroutine's conservatively scanned stack, the stale pointer
// would keep the object alive.
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

// launchArgGoroutine allocates a finalized object and launches a goroutine that
// receives it as an argument, then returns without leaving any reference behind.
// The object reaches the goroutine only through its argument bundle, and the
// only transient copies (of the pointer and the bundle) live in this frame, which
// returns immediately so the later scrubStack recursion reuses and clears it.
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

// testFinishedGoroutineArgs checks that a goroutine which receives a finalized
// object as an argument no longer pins it once finished: the argument bundle the
// goroutine was launched with is dropped when it completes, so the idle collection
// reclaims the object. Without clearing a finished goroutine's args pointer the
// bundle would keep the object alive even after its stack has been zeroed.
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
	testIdleCollect()
	testFinishedGoroutineStacks()
	testFinishedGoroutineArgs()
	println("ok")
}
