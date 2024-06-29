package main

import "unsafe"

//go:wasmimport modulename empty
func empty()

// ERROR: can only use //go:wasmimport on declarations
//
//go:wasmimport modulename implementation
func implementation() {
}

type Uint uint32

type S struct {
	a [4]uint32
	b uintptr
	c int
	d float32
	e float64
}

//go:wasmimport modulename validparam
func validparam(a int32, b uint64, c float64, d unsafe.Pointer, e Uint, f uintptr, g *int32, h *S)

// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type [4]uint32
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type []byte
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type struct{a int}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type chan struct{}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type func()
//
//go:wasmimport modulename invalidparam
func invalidparam(a [4]uint32, b []byte, c struct{ a int }, d chan struct{}, e func())

//go:wasmimport modulename validreturn
func validreturn() int32

// ERROR: //go:wasmimport modulename manyreturns: too many return values
//
//go:wasmimport modulename manyreturns
func manyreturns() (int32, int32)

// ERROR: //go:wasmimport modulename invalidreturn: unsupported result type int
//
//go:wasmimport modulename invalidreturn
func invalidreturn() int

// ERROR: //go:wasmimport modulename invalidUnsafePointerReturn: unsupported result type unsafe.Pointer
//
//go:wasmimport modulename invalidUnsafePointerReturn
func invalidUnsafePointerReturn() unsafe.Pointer
