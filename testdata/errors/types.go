package main

func main() {
	var a int
	a = "foobar"
	nonexisting()
}

// ERROR: # command-line-arguments
// ERROR: types.go:4:6: declared and not used: a
// ERROR: types.go:5:6: cannot use "foobar" (untyped string constant) as int value in assignment
// ERROR: types.go:6:2: undefined: nonexisting
