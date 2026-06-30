package main

import _ "unsafe"

// Creates an external global with name extern_global.
//
//go:extern extern_global
var externGlobal [0]byte

// Creates a
//
//go:align 32
var alignedGlobal [4]uint32

// Test conflicting pragmas (the last one counts).
//
//go:align 64
//go:align 16
var alignedGlobal16 [4]uint32

// Test exported functions.
//
//export extern_func
func externFunc() {
}

// Define a function in a different package using go:linkname.
//
//go:linkname withLinkageName1 somepkg.someFunction1
func withLinkageName1() {
}

// Import a function from a different package using go:linkname.
//
//go:linkname withLinkageName2 somepkg.someFunction2
func withLinkageName2()

// Import a function from a different package using go:linknamestd (the standard
// library variant of go:linkname introduced in Go 1.27).
//
//go:linknamestd withLinkageNameStd somepkg.someFunctionStd
func withLinkageNameStd()

// Function has an 'inline hint', similar to the inline keyword in C.
//
//go:inline
func inlineFunc() {
}

// Function should never be inlined, equivalent to GCC
// __attribute__((noinline)).
//
//go:noinline
func noinlineFunc() {
}

type Int interface {
	int8 | int16
}

// Same for generic functions (but the compiler may miss the pragma due to it
// being generic).
//
//go:noinline
func noinlineGenericFunc[T Int]() {
}

func useGeneric() {
	// Make sure the generic function above is instantiated.
	noinlineGenericFunc[int8]()
}

// This function should have the specified section.
//
//go:section .special_function_section
func functionInSection() {
}

//export exportedFunctionInSection
//go:section .special_function_section
func exportedFunctionInSection() {
}

//go:wasmimport modulename import1
func declaredImport()

// Legacy way of importing a function.
//
//go:wasm-module foobar
//export imported
func foobarImport()

// The wasm-module pragma is not functional here, but it should be safe.
//
//go:wasm-module foobar
//export exported
func foobarExportModule() {
}

// This function should not: it's only a declaration and not a definition.
//
//go:section .special_function_section
func undefinedFunctionNotInSection()

//go:section .special_global_section
var globalInSection uint32

//go:section .special_global_section
//go:extern undefinedGlobalNotInSection
var undefinedGlobalNotInSection uint32

//go:align 1024
//go:section .global_section
var multipleGlobalPragmas uint32

//go:noescape
func doesNotEscapeParam(a *int, b []int, c chan int, d *[0]byte)

// The //go:noescape pragma only works on declarations, not definitions.
//
//go:noescape
func stillEscapes(a *int, b []int, c chan int, d *[0]byte) {
}

//go:noheap
func doesHeapAlloc() *int {
	return new(int)
}

// Define a function in a different package using a file-level go:linkname.
// (Same as withLinkageName1, but with the //go:linkname directive detached
// from the function declaration — see https://github.com/tinygo-org/tinygo/issues/4395)
func withFileLevelLinkageName1() {
}

// Import a function from a different package using a file-level go:linkname.
// (Same as withLinkageName2, but with the //go:linkname directive detached
// from the function declaration.)
func withFileLevelLinkageName2()

//go:linkname withFileLevelLinkageName1 somepkg.someFileLevelFunction1
//go:linkname withFileLevelLinkageName2 somepkg.someFileLevelFunction2

// File-level linkname directives can also appear between two function
// declarations, in which case Go's AST attaches them as the doc comment
// of the following function — even when the directive's localname refers
// to a different function. Exercise that case: the directive below names
// withAdjacentLinkageName, but Go will attach it to
// sentinelAfterAdjacentLinkname's Doc. The file-level scan must find it
// by walking comment groups regardless of which decl they're attached to.
func withAdjacentLinkageName() {
}

//go:linkname withAdjacentLinkageName somepkg.someAdjacentFunction
func sentinelAfterAdjacentLinkname() {
}
