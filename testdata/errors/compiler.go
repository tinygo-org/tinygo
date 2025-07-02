package main

//go:wasmimport foo bar
func foo() {
}

//go:align 7
var global int

// Test for https://github.com/tinygo-org/tinygo/issues/4486
type genericType[T any] struct{}

func (genericType[T]) methodWithoutBody()

func callMethodWithoutBody() {
	msg := &genericType[int]{}
	msg.methodWithoutBody()
}

// ERROR: # command-line-arguments
// ERROR: compiler.go:4:6: can only use //go:wasmimport on declarations
// ERROR: compiler.go:8:5: global variable alignment must be a positive power of two
// ERROR: compiler.go:13:23: missing function body
