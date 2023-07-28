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
