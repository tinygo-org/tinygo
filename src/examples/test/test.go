package main

// This file is here to test features of the Go compiler.

import "unicode"

type Thing struct {
	name string
}

func (t Thing) String() string {
	return t.name
}

type Stringer interface {
	String() string
}

type Foo int

type Number int

func (n Number) Double() int {
	return int(n) * 2
}

type Doubler interface {
	Double() int
}

const SIX = 6

var testmap = map[string]int{"data": 3}

func main() {
	println("Hello world from Go!")
	println("The answer is:", calculateAnswer())
	println("5 ** 2 =", square(5))
	println("3 + 12 =", add(3, 12))
	println("fib(11) =", fib(11))
	println("sumrange(100) =", sumrange(100))
	println("strlen foo:", strlen("foo"))

	// map
	m := map[string]int{"answer": 42, "foo": 3}
	readMap(m, "answer")
	readMap(testmap, "data")

	// slice
	l := 5
	foo := []int{1, 2, 4, 5}
	bar := make([]int, l-2, l)
	println("len/cap foo:", len(foo), cap(foo))
	println("len/cap bar:", len(bar), cap(bar))
	println("len/cap foo[1:2]:", len(foo[1:2]), cap(foo[1:2]))
	println("foo[3]:", foo[3])
	println("sum foo:", sum(foo))
	println("copy foo -> bar:", copy(bar, foo))
	println("sum bar:", sum(bar))

	// interfaces, pointers
	thing := &Thing{"foo"}
	println("thing:", thing.String())
	printItf(5)
	printItf(byte('x'))
	printItf("foo")
	printItf(*thing)
	printItf(thing)
	printItf(Stringer(thing))
	printItf(Number(3))
	s := Stringer(thing)
	println("Stringer.String():", s.String())
	var itf interface{} = s
	println("Stringer.(*Thing).String():", itf.(Stringer).String())

	// unusual calls
	runFunc(hello, 5) // must be indirect to avoid obvious inlining
	testDefer()
	testBound(thing.String)
	func() {
		println("thing inside closure:", thing.String())
	}()
	runFunc(func(i int) {
		println("inside fp closure:", thing.String(), i)
	}, 3)

	// test library functions
	println("lower to upper char:", string('h'), "->", string(unicode.ToUpper('h')))
}

func runFunc(f func(int), arg int) {
	f(arg)
}

func testDefer() {
	i := 1
	defer deferred("...run as defer", i)
	i += 1
	defer deferred("...run as defer", i)
	println("deferring...")
}

func deferred(msg string, i int) {
	println(msg, i)
}

func testBound(f func() string) {
	println("bound method:", f())
}

func readMap(m map[string]int, key string) {
	println("map length:", len(m))
	println("map read:", key, "=", m[key])
	for k, v := range m {
		println(" ", k, "=", v)
	}
}

func hello(n int) {
	println("hello from function pointer:", n)
}

func strlen(s string) int {
	return len(s)
}

func printItf(val interface{}) {
	switch val := val.(type) {
	case Doubler:
		println("is Doubler:", val.Double())
	case int:
		println("is int:", val)
	case byte:
		println("is byte:", val)
	case string:
		println("is string:", val)
	case Thing:
		println("is Thing:", val.String())
	case *Thing:
		println("is *Thing:", val.String())
	case Foo:
		println("is Foo:", val)
	default:
		println("is ?")
	}
}

func calculateAnswer() int {
	seven := 7
	return SIX * seven
}

func square(n int) int {
	return n * n
}

func add(a, b int) int {
	return a + b
}

func fib(n int) int {
	if n <= 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}

func sumrange(n int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		sum += i
	}
	return sum
}

func sum(l []int) int {
	sum := 0
	for _, n := range l {
		sum += n
	}
	return sum
}
