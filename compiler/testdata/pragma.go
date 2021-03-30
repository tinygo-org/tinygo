package main

import _ "unsafe"

// Creates an external global with name extern_global.
//go:extern extern_global
var externGlobal [0]byte

// Creates a
//go:align 32
var alignedGlobal [4]uint32

// Test conflicting pragmas (the last one counts).
//go:align 64
//go:align 16
var alignedGlobal16 [4]uint32

// Test exported functions.
//export extern_func
func externFunc() {
}

// Define a function in a different package using go:linkname.
//go:linkname withLinkageName1 somepkg.someFunction1
func withLinkageName1() {
}

// Import a function from a different package using go:linkname.
//go:linkname withLinkageName2 somepkg.someFunction2
func withLinkageName2()

// Function has an 'inline hint', similar to the inline keyword in C.
//go:inline
func inlineFunc() {
}

// Function should never be inlined, equivalent to GCC
// __attribute__((noinline)).
//go:noinline
func noinlineFunc() {
}
