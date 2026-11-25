//go:build tinygo

package escapeallocs

import (
	"runtime"
	"sync"
	"testing"
)

// The context of an indirect call does not escape through the call. These
// tests check both sides of that rule: closures that are only called while the
// allocating function is running should not allocate, and closures that
// outlive it must keep working after the stack has been reused.

var sinkFunc func() int
var sinkFuncs []func() int
var sinkPtr *int
var sinkAny any

// call calls fn once.
//
//go:noinline
func call(fn func() int) int { return fn() }

// callN calls fn n times.
//
//go:noinline
func callN(n int, fn func(int) int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += fn(i)
	}
	return sum
}

// forward passes fn on to another function.
//
//go:noinline
func forward(fn func() int) int { return call(fn) }

// forwardTwice passes fn on through two more levels.
//
//go:noinline
func forwardTwice(fn func() int) int { return forward(func() int { return fn() + 1 }) }

// ---- closures that must not allocate ----

func TestClosureCalledDirectly(t *testing.T) {
	checkNoAllocs(t, func() {
		x := 7
		sinkInt = call(func() int { return x * 2 })
	})
	x := 7
	if got := call(func() int { return x * 2 }); got != 14 {
		t.Errorf("got %d, want 14", got)
	}
}

func TestClosureCalledInLoop(t *testing.T) {
	checkNoAllocs(t, func() {
		base := 3
		sinkInt = callN(8, func(i int) int { return base + i })
	})
	base := 3
	if got, want := callN(8, func(i int) int { return base + i }), 8*3+28; got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestClosureForwarded(t *testing.T) {
	checkNoAllocs(t, func() {
		x := 5
		sinkInt = forward(func() int { return x })
	})
	x := 5
	if got := forwardTwice(func() int { return x }); got != 6 {
		t.Errorf("got %d, want 6", got)
	}
}

// The captured variables are stored in the context and so stay on the heap
// while the context is a stack slot that is passed to a call. That is
// conservative but correct, so this only checks the result.
func TestClosureCapturingSeveralValues(t *testing.T) {
	a, b, c := 1, 2, 3
	s := "four"
	if got := call(func() int { return a + b + c + len(s) }); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

// A closure in a loop gets a new context in every iteration. The contexts stay
// on the heap, because a stack slot would be reused by the next iteration, but
// every iteration must see its own variable.
func TestClosureCreatedInLoop(t *testing.T) {
	total := 0
	for i := 0; i < 4; i++ {
		total += call(func() int { return i * i })
	}
	if total != 14 {
		t.Errorf("got %d, want 14", total)
	}
}

type counter struct{ n int }

func (c *counter) add(d int) int { c.n += d; return c.n }

// A method value is a closure that holds its receiver in the context.
func TestMethodValue(t *testing.T) {
	checkNoAllocs(t, func() {
		var c counter
		sinkInt = callN(4, c.add)
	})
	var c counter
	if got := callN(4, c.add); got != 1+3+6 || c.n != 6 {
		t.Errorf("got %d (n=%d), want 10 (n=6)", got, c.n)
	}
}

type adder interface{ add(int) int }

//go:noinline
func viaInterface(a adder) int { return a.add(2) + a.add(3) }

// An interface method call is an indirect call whose context is undefined.
func TestInterfaceCall(t *testing.T) {
	checkNoAllocs(t, func() {
		var c counter
		sinkInt = viaInterface(&c)
	})
}

// ---- closures that outlive the allocating function ----

// register keeps fn in a global.
//
//go:noinline
func register(fn func() int) { sinkFunc = fn }

//go:noinline
func registerMany(fn func() int) { sinkFuncs = append(sinkFuncs, fn) }

func TestClosureStoredInGlobal(t *testing.T) {
	func() {
		x, y := 40, 2
		register(func() int { return x + y })
	}()
	clobber(8)
	runtime.GC()
	clobber(8)
	if got := sinkFunc(); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestClosuresAppendedToGlobal(t *testing.T) {
	sinkFuncs = nil
	for i := 0; i < 5; i++ {
		registerMany(func() int { return i * 10 })
	}
	clobber(8)
	runtime.GC()
	for i, fn := range sinkFuncs {
		if got := fn(); got != i*10 {
			t.Errorf("closure %d returned %d, want %d", i, got, i*10)
		}
	}
}

// makeCounter returns a closure with its own state.
//
//go:noinline
func makeCounter(start int) func() int {
	n := start
	return func() int { n++; return n }
}

func TestClosureReturned(t *testing.T) {
	a, b := makeCounter(0), makeCounter(100)
	clobber(8)
	a()
	b()
	clobber(8)
	runtime.GC()
	if x, y := a(), b(); x != 2 || y != 102 {
		t.Errorf("got %d and %d, want 2 and 102", x, y)
	}
}

// identity returns its argument, so the closure flows to the caller again.
//
//go:noinline
func identity(fn func() int) func() int { return fn }

func TestClosureReturnedThroughCallee(t *testing.T) {
	var fn func() int
	func() {
		x := 11
		fn = identity(func() int { return x })
	}()
	clobber(8)
	runtime.GC()
	if got := fn(); got != 11 {
		t.Errorf("got %d, want 11", got)
	}
}

type holder struct{ fn func() int }

//go:noinline
func newHolder(fn func() int) *holder { return &holder{fn} }

func TestClosureStoredInHeapStruct(t *testing.T) {
	var h *holder
	func() {
		x := 21
		h = newHolder(func() int { return x * 2 })
	}()
	clobber(8)
	runtime.GC()
	if got := h.fn(); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

// keep stores fn in a global through one more call, so that the closure is
// captured by a callee that is not the one that gets the context directly.
//
//go:noinline
func keep(fn func() int) { register(func() int { return fn() + 1 }) }

func TestClosureCapturedByWrapper(t *testing.T) {
	func() {
		x := 9
		keep(func() int { return x })
	}()
	clobber(8)
	runtime.GC()
	if got := sinkFunc(); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestClosureConvertedToInterface(t *testing.T) {
	func() {
		x := 13
		sinkAny = func() int { return x }
	}()
	clobber(8)
	runtime.GC()
	if got := sinkAny.(func() int)(); got != 13 {
		t.Errorf("got %d, want 13", got)
	}
}

func TestClosureInGoroutine(t *testing.T) {
	var wg sync.WaitGroup
	results := make([]int, 4)
	for i := range results {
		wg.Add(1)
		x := i * 3
		go func() {
			defer wg.Done()
			clobber(4)
			results[i] = x
		}()
	}
	clobber(8)
	wg.Wait()
	for i, got := range results {
		if got != i*3 {
			t.Errorf("goroutine %d stored %d, want %d", i, got, i*3)
		}
	}
}

func TestClosureInDefer(t *testing.T) {
	got := func() (result int) {
		x := 5
		defer func() { result = x * 2 }()
		x = 6
		clobber(4)
		return 0
	}()
	if got != 12 {
		t.Errorf("got %d, want 12", got)
	}
}

// ---- variables that a closure shares with its caller ----

// leakAddress runs fn, which stores the address of a variable that it
// captured.
//
//go:noinline
func leakAddress(fn func()) { fn() }

// The context can stay on the stack, but the variable that the closure
// exposes to the outside must not.
func TestCapturedVariableAddressEscapes(t *testing.T) {
	func() {
		x := 99
		leakAddress(func() { sinkPtr = &x })
		x = 100
	}()
	clobber(8)
	runtime.GC()
	clobber(8)
	if *sinkPtr != 100 {
		t.Errorf("got %d, want 100", *sinkPtr)
	}
}

// A closure that modifies a captured variable is visible to the caller.
func TestCapturedVariableModified(t *testing.T) {
	x := 0
	callN(5, func(i int) int { x += i; return x })
	if x != 10 {
		t.Errorf("got %d, want 10", x)
	}
	checkNoAllocs(t, func() {
		y := 0
		callN(5, func(i int) int { y += i; return y })
		sinkInt = y
	})
}

// Two closures share one variable, and only one of them is kept.
func TestSharedCapturedVariable(t *testing.T) {
	var get func() int
	func() {
		n := 1
		inc := func() int { n++; return n }
		get = func() int { return n }
		call(inc)
		call(inc)
	}()
	clobber(8)
	runtime.GC()
	if got := get(); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

// ---- recursion ----

//go:noinline
func walk(depth int, fn func(int) int) int {
	if depth == 0 {
		return fn(0)
	}
	return walk(depth-1, func(i int) int { return fn(i) + depth })
}

func TestRecursiveClosureChain(t *testing.T) {
	if got := walk(5, func(i int) int { return i }); got != 15 {
		t.Errorf("got %d, want 15", got)
	}
}

// ---- data arguments of indirect calls ----

// callWithPointer allocates a value and gives it to fn as a data argument,
// which fn may keep. Only the context of an indirect call is exempt from
// escaping.
//
//go:noinline
func callWithPointer(fn func(*int)) {
	x := new(int)
	*x = 77
	fn(x)
}

func TestIndirectCallDataArgumentEscapes(t *testing.T) {
	callWithPointer(func(p *int) { sinkPtr = p })
	clobber(8)
	runtime.GC()
	clobber(8)
	if *sinkPtr != 77 {
		t.Errorf("got %d, want 77", *sinkPtr)
	}
}

type pointerSink interface{ keep(*int) }

type keeper struct{ kept *int }

func (k *keeper) keep(p *int) { k.kept = p }

//go:noinline
func callInterfaceWithPointer(s pointerSink) {
	x := new(int)
	*x = 55
	s.keep(x)
}

// The receiver of an interface method call and its arguments are data, so
// pointers handed to the method must stay alive.
func TestInterfaceCallDataArgumentEscapes(t *testing.T) {
	k := &keeper{}
	callInterfaceWithPointer(k)
	clobber(8)
	runtime.GC()
	clobber(8)
	if *k.kept != 55 {
		t.Errorf("got %d, want 55", *k.kept)
	}
}
