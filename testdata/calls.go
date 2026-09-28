package main

import "runtime/volatile"

type Thing struct {
	name string
}

type ThingOption func(*Thing)

func WithName(name string) ThingOption {
	return func(t *Thing) {
		t.name = name
	}
}

func NewThing(opts ...ThingOption) *Thing {
	t := &Thing{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t Thing) String() string {
	return t.name
}

func (t Thing) Print(arg string) {
	println("Thing.Print:", t.name, "arg:", arg)
}

type Printer interface {
	Print(string)
}

func main() {
	thing := &Thing{"foo"}

	// function pointers
	runFunc(hello, 5) // must be indirect to avoid obvious inlining

	// deferred functions
	testDefer()

	// defers in loop
	testDeferLoop()

	//defer func variable call
	testDeferFuncVar()

	//More complicated func variable call
	testMultiFuncVar()

	// Take a bound method and use it as a function pointer.
	// This function pointer needs a context pointer.
	testBound(thing.String)

	// closures
	func() {
		println("thing inside closure:", thing.String())
	}()
	runFunc(func(i int) {
		println("inside fp closure:", thing.String(), i)
	}, 3)

	// functional arguments
	thingFunctionalArgs1 := NewThing()
	thingFunctionalArgs1.Print("functional args 1")
	thingFunctionalArgs2 := NewThing(WithName("named thing"))
	thingFunctionalArgs2.Print("functional args 2")

	// regression testing
	regression1033()
	regression5541()

	//Test deferred builtins
	testDeferBuiltinClose()
	testDeferBuiltinDelete()

	// Check for issue 1304.
	// There are two fields in this struct, one of which is zero-length so the
	// other covers the entire struct. This led to a verification error for
	// debug info, which used DW_OP_LLVM_fragment for a field that practically
	// covered the entire variable.
	var x issue1304
	x.call()

	if testDeferElse(false) != 0 {
		println("else defer returned wrong value")
	}

	// See https://github.com/tinygo-org/tinygo/pull/5776.
	err := multiReturnEscape()
	clobberStack()
	println("multiple return values:", err.(*multiReturnErr).x[1])
	err = singleArrayEscape()
	clobberStack()
	println("single array return:", err.(*multiReturnErr).x[1])
	err = forwardedReturnEscape()
	clobberStack()
	println("forwarded return:", err.(*multiReturnErr).x[1])
	s := sliceReturnEscape()
	clobberStack()
	println("slice return:", s[1])
	globalReturnEscape()
	clobberStack()
	println("global escape:", globalEscaped.(*multiReturnErr).x[1])
	fn := closureReturnEscape()
	clobberStack()
	println("closure escape:", fn())
	err = phiReturnEscape(false)
	clobberStack()
	println("phi escape:", err.(*multiReturnErr).x[1])
	g := interfaceReturnEscape()
	clobberStack()
	println("interface return:", g.get())
	err = deepForwardEscape()
	clobberStack()
	println("deep forward:", err.(*multiReturnErr).x[1])
}

type multiReturnErr struct{ x [4]int }

func (e *multiReturnErr) Error() string { return "multiReturnErr" }

//go:noinline
func wrapMultiReturn(e *multiReturnErr) (int, error) { return 1, e }

//go:noinline
func multiReturnEscape() error {
	e := &multiReturnErr{}
	e.x[1] = 42
	_, err := wrapMultiReturn(e)
	return err
}

//go:noinline
func wrapSingleArrayReturn(e *multiReturnErr) [1]error { return [1]error{e} }

//go:noinline
func singleArrayEscape() error {
	e := &multiReturnErr{}
	e.x[1] = 43
	a := wrapSingleArrayReturn(e)
	return a[0]
}

//go:noinline
func forwardMultiReturn(e *multiReturnErr) error {
	_, err := wrapMultiReturn(e)
	return err
}

//go:noinline
func forwardedReturnEscape() error {
	e := &multiReturnErr{}
	e.x[1] = 44
	return forwardMultiReturn(e)
}

//go:noinline
func wrapSliceReturn(b []int) (int, []int) { return 1, b }

// The allocation reaches the caller inside the {ptr, len, cap} slice aggregate
// of a multiple return value.
//
//go:noinline
func sliceReturnEscape() []int {
	b := make([]int, 4)
	b[1] = 45
	_, s := wrapSliceReturn(b)
	return s
}

var globalEscaped error

// The allocation escapes to a package-level variable rather than through a
// return.
//
//go:noinline
func globalReturnEscape() {
	e := &multiReturnErr{}
	e.x[1] = 46
	_, err := wrapMultiReturn(e)
	globalEscaped = err
}

// The allocation escapes by being captured in a closure.
//
//go:noinline
func closureReturnEscape() func() int {
	e := &multiReturnErr{}
	e.x[1] = 47
	_, err := wrapMultiReturn(e)
	return func() int { return err.(*multiReturnErr).x[1] }
}

// The extracted value reaches the return through a phi.
//
//go:noinline
func phiReturnEscape(cond bool) error {
	e := &multiReturnErr{}
	e.x[1] = 48
	_, err := wrapMultiReturn(e)
	if cond {
		return nil
	}
	return err
}

type multiReturnGetter interface{ get() int }

func (e *multiReturnErr) get() int { return e.x[1] }

//go:noinline
func wrapInterfaceReturn(e *multiReturnErr) (int, multiReturnGetter) { return 1, e }

// The allocation reaches the caller inside a non-error interface value and is
// then called through.
//
//go:noinline
func interfaceReturnEscape() multiReturnGetter {
	e := &multiReturnErr{}
	e.x[1] = 49
	_, g := wrapInterfaceReturn(e)
	return g
}

//go:noinline
func deepForward1(e *multiReturnErr) (error, int) {
	_, err := wrapMultiReturn(e)
	return err, 2
}

//go:noinline
func deepForward2(e *multiReturnErr) [1]error {
	err, _ := deepForward1(e)
	return [1]error{err}
}

//go:noinline
func deepForward3(e *multiReturnErr) error {
	a := deepForward2(e)
	return a[0]
}

// The allocation is repacked into a different aggregate shape at each of four
// call levels before reaching the caller.
//
//go:noinline
func deepForwardEscape() error {
	e := &multiReturnErr{}
	e.x[1] = 50
	return deepForward3(e)
}

//go:noinline
func clobberStack() {
	var a [32]uint32
	for i := range a {
		// Keep the stack writes. See https://llvm.org/docs/LangRef.html#volatile-memory-accesses.
		volatile.StoreUint32(&a[i], uint32(1000+i))
	}
}

func runFunc(f func(int), arg int) {
	f(arg)
}

func hello(n int) {
	println("hello from function pointer:", n)
}

func testDefer() {
	defer exportedDefer()

	i := 1
	defer deferred("...run as defer", i)
	i++
	defer func() {
		println("...run closure deferred:", i)
	}()
	i++
	defer deferred("...run as defer", i)
	i++

	var t Printer = &Thing{"foo"}
	defer t.Print("bar")

	println("deferring...")
	d := dumb{}
	defer d.Value(0)
}

func testDeferLoop() {
	for j := 0; j < 4; j++ {
		defer deferred("loop", j)
	}
}

func testDeferFuncVar() {
	dummy, f := deferFunc()
	dummy++
	defer f(1)
}

func testMultiFuncVar() {
	f := multiFuncDefer()
	defer f(1)
}

func testDeferBuiltinClose() {
	i := make(chan int)
	func() {
		defer close(i)
	}()
	if n, ok := <-i; n != 0 || ok {
		println("expected to read 0 from closed channel")
	}
}

func testDeferBuiltinDelete() {
	m := map[int]int{3: 30, 5: 50}
	func() {
		defer delete(m, 3)
		if m[3] != 30 {
			println("expected m[3] to be 30")
		}
	}()
	if m[3] != 0 {
		println("expected m[3] to be 0")
	}
}

type dumb struct {
}

func (*dumb) Value(key interface{}) interface{} {
	return nil
}

func deferred(msg string, i int) {
	println(msg, i)
}

//export __exportedDefer
func exportedDefer() {
	println("...exported defer")
}

func deferFunc() (int, func(int)) {
	return 0, func(i int) { println("...extracted defer func ", i) }
}

func multiFuncDefer() func(int) {
	i := 0

	if i > 0 {
		return func(i int) { println("Should not have gotten here. i = ", i) }
	}

	return func(i int) { println("Called the correct function. i = ", i) }
}

func testBound(f func() string) {
	println("bound method:", f())
}

// regression1033 is a regression test for https://github.com/tinygo-org/tinygo/issues/1033.
// In previous versions of the compiler, a deferred call to an interface would create an instruction that did not dominate its uses.
func regression1033() {
	foo(&Bar{})
}

type Bar struct {
	empty bool
}

func (b *Bar) Close() error {
	return nil
}

type Closer interface {
	Close() error
}

func foo(bar *Bar) error {
	var a int
	if !bar.empty {
		a = 10
		if a != 5 {
			return nil
		}
	}

	var c Closer = bar
	defer c.Close()

	return nil
}

var regression5541RuntimeClosed, regression5541ModuleClosed int

type regression5541Closer interface {
	Close()
}

type regression5541Runtime interface {
	Foo()
	regression5541Closer
}

type regression5541RuntimeImpl struct{}

func (*regression5541RuntimeImpl) Foo() {
}

func (*regression5541RuntimeImpl) Close() {
	regression5541RuntimeClosed++
}

type regression5541Module struct{}

func (*regression5541Module) Close() {
	regression5541ModuleClosed++
}

func regression5541() {
	func() {
		var runtime regression5541Runtime = &regression5541RuntimeImpl{}
		defer runtime.Close()

		var module regression5541Closer = &regression5541Module{}
		defer module.Close()
	}()

	if regression5541RuntimeClosed != 1 || regression5541ModuleClosed != 1 {
		println("deferred interface methods were not both called")
	}
}

type issue1304 struct {
	a [0]int // zero-length field
	b int    // field 'b' covers entire struct
}

func (x issue1304) call() {
	// nothing to do
}

type recursiveFuncType func(recursiveFuncType)

var recursiveFunction recursiveFuncType

//go:noinline
func testDeferElse(b bool) (r int) {
	if !b {
		defer func() {
			r = 0

		}()
	}

	return 1
}
