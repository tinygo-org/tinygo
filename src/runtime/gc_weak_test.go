//go:build gc.boehm || gc.conservative || gc.precise

package runtime_test

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"weak"
)

type weakTestObj struct {
	a int
	b int
}

var weakScrubSink int

//go:noinline
func scrubWeakStack(depth int) int {
	if depth <= 0 {
		return weakScrubSink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	weakScrubSink += buf[depth&63]
	return scrubWeakStack(depth-1) + buf[0]
}

//go:noinline
func makeDroppedWeak() (weak.Pointer[weakTestObj], weak.Pointer[weakTestObj], weak.Pointer[int]) {
	obj := &weakTestObj{a: 1, b: 2}
	return weak.Make(obj), weak.Make(obj), weak.Make(&obj.a)
}

//go:noinline
func makeDroppedWeakWithFinalizer(ran *atomic.Bool, sawNil *atomic.Bool) weak.Pointer[weakTestObj] {
	obj := &weakTestObj{a: 3, b: 4}
	wp := weak.Make(obj)
	runtime.SetFinalizer(obj, func(*weakTestObj) {
		if wp.Value() == nil {
			sawNil.Store(true)
		}
		ran.Store(true)
	})
	return wp
}

func TestWeakPointer(t *testing.T) {
	var zero weak.Pointer[weakTestObj]
	if zero.Value() != nil {
		t.Fatal("zero weak.Pointer is not nil")
	}
	if weak.Make[weakTestObj](nil) != zero {
		t.Fatal("weak.Make(nil) != zero")
	}

	w1, w2, wField := makeDroppedWeak()
	if w1 != w2 {
		t.Fatal("weak.Make on same pointer produced unequal handles")
	}
	other := &weakTestObj{a: 9}
	wOther := weak.Make(other)
	if w1 == wOther {
		t.Fatal("weak.Make on distinct pointers produced equal handles")
	}
	if wOther.Value() != other {
		t.Fatal("live weak pointer returned wrong value")
	}

	scrubWeakStack(16)
	runtime.GC()
	runtime.GC()

	if v := w1.Value(); v != nil {
		t.Fatalf("weak pointer not cleared after GC: %p", v)
	}
	if v := wField.Value(); v != nil {
		t.Fatalf("interior weak pointer not cleared after GC: %p", v)
	}
	if w1 != w2 {
		t.Fatal("equal weak pointers became unequal after GC")
	}
	if w1 == wOther {
		t.Fatal("distinct weak pointers became equal after GC")
	}
	if wOther.Value() != other {
		t.Fatal("live weak pointer cleared prematurely")
	}
	runtime.KeepAlive(other)
}

func TestWeakPointerFinalizer(t *testing.T) {
	var ran, sawNil atomic.Bool
	wp := makeDroppedWeakWithFinalizer(&ran, &sawNil)

	scrubWeakStack(16)
	runtime.GC()
	if wp.Value() != nil {
		t.Fatal("weak pointer not cleared when finalizer was enqueued")
	}

	deadline := time.Now().Add(2 * time.Second)
	for !ran.Load() && time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	if !ran.Load() {
		t.Fatal("finalizer did not run")
	}
	if !sawNil.Load() {
		t.Fatal("weak pointer was non-nil inside finalizer")
	}
	runtime.GC()
	if wp.Value() != nil {
		t.Fatal("weak pointer non-nil after finalizer ran")
	}
}
