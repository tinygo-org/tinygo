package main

func main() {
	testIntegerRange()
	testLoopVar()
}

func testIntegerRange() {
	for i := range 10 {
		println(10 - i)
	}
	println("go1.22 has lift-off!")
}

func testLoopVar() {
	var f func() int
	for i := 0; i < 1; i++ {
		if i == 0 {
			f = func() int { return i }
		}
	}
	// Prints 1 in Go 1.21, or 0 in Go 1.22.
	// TODO: this still prints Go 1.21 even in Go 1.22. We probably need to
	// specify the Go version somewhere.
	n := f()
	if n == 0 {
		println("behaves like Go 1.22")
	} else if n == 1 {
		println("behaves like Go 1.21")
	} else {
		println("unknown behavior")
	}
}
