package main

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

type issue1304 struct {
	a [0]int // zero-length field
	b int    // field 'b' covers entire struct
}

func (x issue1304) call() {
	// nothing to do
}

type recursiveFuncType func(recursiveFuncType)

var recursiveFunction recursiveFuncType
