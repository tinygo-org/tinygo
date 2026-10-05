package main

var mode string

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

func inlinePanic() int {
	panic("boom")
}

func direct() {
	defer func() {}()
	_ = inlinePanic()
}

func main() {
	if mode == "direct" {
		direct()
		return
	}
	defer func() {}()
	outer()
}
