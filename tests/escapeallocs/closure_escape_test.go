//go:build tinygo

package escapeallocs

import (
	"context"
	"runtime"
	"testing"
	"time"
)

// These tests look for places where an allocation is moved to the stack even
// though it outlives the function that allocates it. Each one hides the call
// target from the compiler, so that the call stays an indirect call (a
// constant func value would be turned into a direct call before the escape
// analysis runs), lets the allocating function return, overwrites the stack
// that it used, and only then looks at the data. An allocation that wrongly
// lives on the dead stack frame reads back as 0xAA bytes or crashes.

// runtimeZero is zero, but the compiler cannot know that.
var runtimeZero = int(time.Now().UnixNano() >> 62)

// opaque returns a or b, depending on a value that is only known at run time.
//
//go:noinline
func opaque(a, b func() int) func() int {
	if runtimeZero != 0 {
		return b
	}
	return a
}

//go:noinline
func opaquePtr(a, b func(*int)) func(*int) {
	if runtimeZero != 0 {
		return b
	}
	return a
}

// settle lets everything that should be dead die and reuses the stack.
func settle() {
	clobber(16)
	runtime.GC()
	clobber(16)
}

func noop() int { return 0 }

func noopPtr(*int) {}

// ---- data arguments of calls that cannot be resolved ----

//go:noinline
func callOpaqueWithPointer(fn func(*int)) {
	x := new(int)
	*x = 77
	fn(x)
}

var keptInt *int

func TestOpaqueCallbackKeepsDataArgument(t *testing.T) {
	callOpaqueWithPointer(opaquePtr(func(p *int) { keptInt = p }, noopPtr))
	settle()
	if *keptInt != 77 {
		t.Errorf("got %d, want 77", *keptInt)
	}
}

// The pointer is both the data argument and part of the context.
//
//go:noinline
func callOpaqueWithPointerAndContext(fn func(*int)) {
	x := new(int)
	*x = 88
	fn(x)
}

func TestOpaqueCallbackKeepsDataArgumentAndCapturesIt(t *testing.T) {
	var all []*int
	callOpaqueWithPointerAndContext(opaquePtr(func(p *int) { all = append(all, p) }, noopPtr))
	settle()
	if len(all) != 1 || *all[0] != 88 {
		t.Errorf("got %v, want one pointer to 88", all)
	}
}

type pointerKeeper interface{ keep(*int) }

type keepsPointer struct{ kept *int }

func (k *keepsPointer) keep(p *int) { k.kept = p }

type dropsPointer struct{}

func (dropsPointer) keep(*int) {}

//go:noinline
func chooseKeeper(k *keepsPointer) pointerKeeper {
	if runtimeZero != 0 {
		return dropsPointer{}
	}
	return k
}

//go:noinline
func callKeeper(k pointerKeeper) {
	x := new(int)
	*x = 55
	k.keep(x)
}

func TestInterfaceMethodKeepsDataArgument(t *testing.T) {
	k := &keepsPointer{}
	callKeeper(chooseKeeper(k))
	settle()
	if *k.kept != 55 {
		t.Errorf("got %d, want 55", *k.kept)
	}
}

// ---- closures that are kept by something other than a plain store ----

var (
	keptFunc  func() int
	keptFuncs []func() int
)

//go:noinline
func sendClosure(ch chan func() int, fn func() int) { ch <- fn }

func TestClosureSentOnChannel(t *testing.T) {
	ch := make(chan func() int, 1)
	func() {
		x := 31
		sendClosure(ch, opaque(func() int { return x }, noop))
	}()
	settle()
	if got := (<-ch)(); got != 31 {
		t.Errorf("got %d, want 31", got)
	}
}

//go:noinline
func putInMap(m map[string]func() int, fn func() int) { m["fn"] = fn }

func TestClosureStoredInMap(t *testing.T) {
	m := map[string]func() int{}
	func() {
		x := 32
		putInMap(m, opaque(func() int { return x }, noop))
	}()
	settle()
	if got := m["fn"](); got != 32 {
		t.Errorf("got %d, want 32", got)
	}
}

type closureBox struct {
	name string
	fn   func() int
}

//go:noinline
func boxClosure(fn func() int) closureBox { return closureBox{"box", fn} }

func TestClosureReturnedInsideStruct(t *testing.T) {
	var box closureBox
	func() {
		x := 33
		box = boxClosure(opaque(func() int { return x }, noop))
	}()
	settle()
	if got := box.fn(); got != 33 {
		t.Errorf("got %d, want 33", got)
	}
}

//go:noinline
func boxInInterface(fn func() int) any { return closureBox{"box", fn} }

func TestClosureInStructInInterface(t *testing.T) {
	var v any
	func() {
		x := 34
		v = boxInInterface(opaque(func() int { return x }, noop))
	}()
	settle()
	if got := v.(closureBox).fn(); got != 34 {
		t.Errorf("got %d, want 34", got)
	}
}

// keepIfTrue only keeps the closure on one path. The analysis has to assume
// that path is taken.
//
//go:noinline
func keepIfTrue(fn func() int, keep bool) int {
	if keep {
		keptFunc = fn
	}
	return fn()
}

func TestClosureKeptOnOnePath(t *testing.T) {
	keptFunc = nil
	func() {
		x := 35
		keepIfTrue(opaque(func() int { return x }, noop), runtimeZero == 0)
	}()
	settle()
	if got := keptFunc(); got != 35 {
		t.Errorf("got %d, want 35", got)
	}
}

// keepAfterCall calls the closure first and keeps it afterwards.
//
//go:noinline
func keepAfterCall(fn func() int) int {
	r := fn()
	keptFuncs = append(keptFuncs, fn)
	return r
}

func TestClosureKeptAfterBeingCalled(t *testing.T) {
	keptFuncs = nil
	func() {
		x := 36
		keepAfterCall(opaque(func() int { return x }, noop))
	}()
	settle()
	if got := keptFuncs[0](); got != 36 {
		t.Errorf("got %d, want 36", got)
	}
}

// wrapAndKeep keeps a closure that calls the one it was given.
//
//go:noinline
func wrapAndKeep(fn func() int) {
	keptFunc = func() int { return fn() * 2 }
}

func TestClosureCapturedByAnotherClosure(t *testing.T) {
	func() {
		x := 20
		wrapAndKeep(opaque(func() int { return x + 1 }, noop))
	}()
	settle()
	if got := keptFunc(); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

// A closure that refers to itself through a captured variable.
func TestSelfReferencingClosure(t *testing.T) {
	var fact func(int) int
	func() {
		fact = func(n int) int {
			if n <= 1 {
				return 1
			}
			return n * fact(n-1)
		}
	}()
	settle()
	if got := fact(5); got != 120 {
		t.Errorf("got %d, want 120", got)
	}
}

//go:noinline
func makeSelfReferencing() func(int) int {
	var f func(int) int
	f = func(n int) int {
		if n <= 0 {
			return 0
		}
		return n + f(n-1)
	}
	return f
}

func TestReturnedSelfReferencingClosure(t *testing.T) {
	f := makeSelfReferencing()
	settle()
	if got := f(4); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

// ---- things that call a closure after the creator is gone ----

func TestClosureKeptByTestCleanup(t *testing.T) {
	cleaned := 0
	t.Run("register", func(t *testing.T) {
		x := 41
		fn := opaque(func() int { return x }, noop)
		t.Cleanup(func() { cleaned = fn() })
	})
	if cleaned != 41 {
		t.Errorf("cleanup saw %d, want 41", cleaned)
	}
}

func TestClosureKeptByAfterFunc(t *testing.T) {
	done := make(chan int, 1)
	func() {
		x := 42
		fn := opaque(func() int { return x }, noop)
		time.AfterFunc(time.Millisecond, func() {
			clobber(8)
			done <- fn()
		})
	}()
	settle()
	select {
	case got := <-done:
		if got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timer did not fire")
	}
}

func TestClosureKeptByContextAfterFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan int, 1)
	func() {
		x := 43
		fn := opaque(func() int { return x }, noop)
		context.AfterFunc(ctx, func() {
			clobber(8)
			done <- fn()
		})
	}()
	settle()
	cancel()
	select {
	case got := <-done:
		if got != 43 {
			t.Errorf("got %d, want 43", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("callback did not run")
	}
}

//go:noinline
func startAfterReturn(fn func() int, out chan<- int) {
	go func() {
		clobber(8)
		out <- fn()
	}()
}

func TestClosureUsedByGoroutineAfterReturn(t *testing.T) {
	out := make(chan int)
	func() {
		x := 44
		startAfterReturn(opaque(func() int { return x }, noop), out)
	}()
	settle()
	if got := <-out; got != 44 {
		t.Errorf("got %d, want 44", got)
	}
}

// ---- bound method values ----

type big struct {
	a, b, c, d int
}

func (b *big) sum() int { return b.a + b.b + b.c + b.d }

func (b big) total() int { return b.a + b.b + b.c + b.d }

//go:noinline
func boundSum(b *big) func() int { return b.sum }

//go:noinline
func boundTotal(b big) func() int { return b.total }

// The method value of a pointer receiver keeps the receiver alive.
func TestBoundMethodKeepsReceiver(t *testing.T) {
	var fn func() int
	func() {
		b := &big{1, 2, 3, 4}
		fn = boundSum(b)
		b.d = 14
	}()
	settle()
	if got := fn(); got != 20 {
		t.Errorf("got %d, want 20", got)
	}
}

// The method value of a value receiver holds a copy.
func TestBoundMethodCopiesValueReceiver(t *testing.T) {
	var fn func() int
	func() {
		b := big{1, 2, 3, 4}
		fn = boundTotal(b)
		b.d = 100
	}()
	settle()
	if got := fn(); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

// ---- variables that closures expose ----

var keptPtrs []*int

// Every iteration's variable has to survive, even though the context of the
// closure that is called in the loop can stay local.
func TestCapturedVariablesOfSeveralIterationsEscape(t *testing.T) {
	keptPtrs = nil
	for i := 0; i < 4; i++ {
		x := i * 10
		callOpaqueVoid(opaqueVoid(func() { keptPtrs = append(keptPtrs, &x) }, noopVoid))
		x++
	}
	settle()
	for i, p := range keptPtrs {
		if *p != i*10+1 {
			t.Errorf("variable %d is %d, want %d", i, *p, i*10+1)
		}
	}
}

func noopVoid() {}

//go:noinline
func opaqueVoid(a, b func()) func() {
	if runtimeZero != 0 {
		return b
	}
	return a
}

//go:noinline
func callOpaqueVoid(fn func()) { fn() }

// A closure that returns a pointer to a variable it captured.
//
//go:noinline
func callAndKeepResult(fn func() *int) { keptInt = fn() }

func TestClosureReturnsCapturedVariable(t *testing.T) {
	func() {
		x := 66
		callAndKeepResult(func() *int { return &x })
		x = 67
	}()
	settle()
	if *keptInt != 67 {
		t.Errorf("got %d, want 67", *keptInt)
	}
}

// ---- a loop that reuses one stack slot ----

func TestClosuresFromLoopKeptInOtherFunction(t *testing.T) {
	keptFuncs = nil
	for i := 0; i < 6; i++ {
		x := i + 100
		keepAfterCall(opaque(func() int { return x }, noop))
	}
	settle()
	for i, fn := range keptFuncs {
		if got := fn(); got != i+100 {
			t.Errorf("closure %d returned %d, want %d", i, got, i+100)
		}
	}
}
