
package main

type Thing struct {
	name string
}

func (t Thing) String() string {
	return t.name
}

const SIX = 6

func main() {
	println("Hello world from Go!")
	println("The answer is:", calculateAnswer())
	println("5 ** 2 =", square(5))
	println("3 + 12 =", add(3, 12))
	println("fib(11) =", fib(11))
	println("sumrange(100) =", sumrange(100))

	thing := Thing{"foo"}
	println("thing:", thing.String())
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
