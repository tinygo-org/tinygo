package main

//go:wasmimport foo bar
func foo() {
}

//go:align 7
var global int

// ERROR: # command-line-arguments
// ERROR: compiler.go:4:6: can only use //go:wasmimport on declarations
// ERROR: compiler.go:8:5: global variable alignment must be a positive power of two
