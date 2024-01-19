package main

// This package verifies that the Go language version is correctly picked up
// from the go.mod file.

func main() {
	testLoopVar()
}

func testLoopVar() {
	var f func() int
	for i := 0; i < 1; i++ {
		if i == 0 {
			f = func() int { return i }
		}
	}
	// Variable n is 1 in Go 1.21, or 0 in Go 1.22.
	n := f()
	if n == 0 {
		println("loops behave like Go 1.22")
	} else if n == 1 {
		println("loops behave like Go 1.21")
	} else {
		println("unknown loop behavior")
	}
}
