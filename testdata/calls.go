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
}

func runFunc(f func(int), arg int) {
	f(arg)
}

func hello(n int) {
	println("hello from function pointer:", n)
}

func testDefer() {
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
}

func testDeferLoop() {
	for j := 0; j < 4; j++ {
		defer deferred("loop", j)
	}
}

func deferred(msg string, i int) {
	println(msg, i)
}

func testBound(f func() string) {
	println("bound method:", f())
}
