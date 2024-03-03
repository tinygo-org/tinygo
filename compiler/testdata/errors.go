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

//go:wasmimport modulename validparam
func validparam(a int32, b uint64, c float64, d unsafe.Pointer, e Uint)

// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type int
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type string
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type []byte
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type *int32
//
//go:wasmimport modulename invalidparam
func invalidparam(a int, b string, c []byte, d *int32)

//go:wasmimport modulename validreturn
func validreturn() int32

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
