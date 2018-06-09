
package main

type Thing struct {
	name string
}

func (t Thing) String() string {
	return t.name
}

type Stringer interface {
	String() string
}

const SIX = 6

func main() {
	println("Hello world from Go!")
	println("The answer is:", calculateAnswer())
	println("5 ** 2 =", square(5))
	println("3 + 12 =", add(3, 12))
	println("fib(11) =", fib(11))
	println("sumrange(100) =", sumrange(100))
	println("strlen foo:", strlen("foo"))

	thing := &Thing{"foo"}
	println("thing:", thing.String())
	printItf(5)
	printItf(byte('x'))
	printItf("foo")
	printItf(*thing)
	printItf(thing)
	printItf(Stringer(thing))
	s := Stringer(thing)
	println("Stringer.String():", s.String())

	runFunc(hello, 5) // must be indirect to avoid obvious inlining
}

func runFunc(f func(int), arg int) {
	f(arg)
}

func hello(n int) {
	println("hello from function pointer:", n)
}

func strlen(s string) int {
	return len(s)
}

func printItf(val interface{}) {
	switch val := val.(type) {
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
	return fib(n - 1) + fib(n - 2)
}

func sumrange(n int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		sum += i
	}
	return sum
}
