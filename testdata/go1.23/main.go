package main

func main() {
	testFuncRange(counter)
}

func testFuncRange(f func(yield func(int) bool)) {
	for i := range f {
		println(i)
	}
	println("go1.23 has lift-off!")
}

func counter(yield func(int) bool) {
	for i := 10; i >= 1; i-- {
		yield(i)
	}
}
