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
func validparam(a int32, b uint64, c float64, d unsafe.Pointer, e Uint, f uintptr, g string, h *int32, i *S)

// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type [4]uint32
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type []byte
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type struct{a int}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type chan struct{}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type func()
//
//go:wasmimport modulename invalidparam
func invalidparam(a [4]uint32, b []byte, c struct{ a int }, d chan struct{}, e func())

//go:wasmimport modulename validreturn_int32
func validreturn_int32() int32

//go:wasmimport modulename validreturn_int
func validreturn_int() int

//go:wasmimport modulename validreturn_ptr_int32
func validreturn_ptr_int32() *int32

//go:wasmimport modulename validreturn_ptr_string
func validreturn_ptr_string() *string

//go:wasmimport modulename validreturn_ptr_struct
func validreturn_ptr_struct() *S

//go:wasmimport modulename validreturn_unsafe_pointer
func validreturn_unsafe_pointer() unsafe.Pointer

// ERROR: //go:wasmimport modulename manyreturns: too many return values
//
//go:wasmimport modulename manyreturns
func manyreturns() (int32, int32)

// ERROR: //go:wasmimport modulename invalidreturn_func: unsupported result type func()
//
//go:wasmimport modulename invalidreturn_func
func invalidreturn_func() func()

// ERROR: //go:wasmimport modulename invalidreturn_slice_byte: unsupported result type []byte
//
//go:wasmimport modulename invalidreturn_slice_byte
func invalidreturn_slice_byte() []byte

// ERROR: //go:wasmimport modulename invalidreturn_chan_int: unsupported result type chan int
//
//go:wasmimport modulename invalidreturn_chan_int
func invalidreturn_chan_int() chan int

// ERROR: //go:wasmimport modulename invalidreturn_string: unsupported result type string
//
//go:wasmimport modulename invalidreturn_string
func invalidreturn_string() string
