package main

//go:noinline
func panicHere() {
	panic("boom")
}

var panicFunc = panicHere

//go:noinline
func inner() {
	defer func() {}()
	panicFunc()
}

//go:noinline
func outer() {
	defer func() {}()
	inner()
}

func main() {
	defer func() {}()
	outer()
}
