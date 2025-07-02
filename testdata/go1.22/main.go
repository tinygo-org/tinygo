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
