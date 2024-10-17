package main

import "iter"

func main() {
	testFuncRange(counter)
	testIterPull(counter)
	println("go1.23 has lift-off!")
}

func testFuncRange(it iter.Seq[int]) {
	for i := range it {
		println(i)
	}
}

func testIterPull(it iter.Seq[int]) {
	next, stop := iter.Pull(it)
	defer stop()
	for {
		i, ok := next()
		if !ok {
			break
		}
		println(i)
	}
}

func counter(yield func(int) bool) {
	for i := 10; i >= 1; i-- {
		yield(i)
	}
}
