package main

func init() {
	println("called init")
}

//go:wasmimport tester callTestMain
func callTestMain()

func main() {
	// main.main is not used when using -buildmode=c-shared.
	callTestMain()
}

//go:wasmexport hello
func hello() {
	println("hello!")
}

//go:wasmexport add
func add(a, b int) int {
	println("called add:", a, b)
	return a + b
}

//go:wasmimport tester callOutside
func callOutside(a, b int) int

//go:wasmexport reentrantCall
func reentrantCall(a, b int) int {
	println("reentrantCall:", a, b)
	result := callOutside(a, b)
	println("reentrantCall result:", result)
	return result
}
