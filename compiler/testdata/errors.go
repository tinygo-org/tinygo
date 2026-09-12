package main

import (
	"structs"
	"unsafe"
)

//go:wasmimport modulename empty
func empty()

// ERROR: can only use //go:wasmimport on declarations
//
//go:wasmimport modulename implementation
func implementation() {
}

type Uint uint32

type S struct {
	_ structs.HostLayout
	a [4]uint32
	b uintptr
	d float32
	e float64
}

//go:wasmimport modulename validparam
func validparam(a int32, b uint64, c float64, d unsafe.Pointer, e Uint, f uintptr, g string, h *int32, i *S, j *struct{}, k *[8]uint8)

// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type [4]uint32
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type []byte
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type struct{a int}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type chan struct{}
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type func()
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type int
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type uint
// ERROR: //go:wasmimport modulename invalidparam: unsupported parameter type [8]int
//
//go:wasmimport modulename invalidparam
func invalidparam(a [4]uint32, b []byte, c struct{ a int }, d chan struct{}, e func(), f int, g uint, h [8]int)

// ERROR: //go:wasmimport modulename invalidparam_no_hostlayout: unsupported parameter type *struct{int}
// ERROR: //go:wasmimport modulename invalidparam_no_hostlayout: unsupported parameter type *struct{string}
//
//go:wasmimport modulename invalidparam_no_hostlayout
func invalidparam_no_hostlayout(a *struct{ int }, b *struct{ string })

//go:wasmimport modulename validreturn_int32
func validreturn_int32() int32

//go:wasmimport modulename validreturn_ptr_int32
func validreturn_ptr_int32() *int32

//go:wasmimport modulename validreturn_ptr_string
func validreturn_ptr_string() *string

//go:wasmimport modulename validreturn_ptr_struct
func validreturn_ptr_struct() *S

//go:wasmimport modulename validreturn_ptr_struct
func validreturn_ptr_empty_struct() *struct{}

//go:wasmimport modulename validreturn_ptr_array
func validreturn_ptr_array() *[8]uint8

//go:wasmimport modulename validreturn_unsafe_pointer
func validreturn_unsafe_pointer() unsafe.Pointer

// ERROR: //go:wasmimport modulename manyreturns: too many return values
//
//go:wasmimport modulename manyreturns
func manyreturns() (int32, int32)

// ERROR: //go:wasmimport modulename invalidreturn_int: unsupported result type int
//
//go:wasmimport modulename invalidreturn_int
func invalidreturn_int() int

// ERROR: //go:wasmimport modulename invalidreturn_int: unsupported result type uint
//
//go:wasmimport modulename invalidreturn_int
func invalidreturn_uint() uint

// ERROR: //go:wasmimport modulename invalidreturn_func: unsupported result type func()
//
//go:wasmimport modulename invalidreturn_func
func invalidreturn_func() func()

// ERROR: //go:wasmimport modulename invalidreturn_pointer_array_int: unsupported result type *[8]int
//
//go:wasmimport modulename invalidreturn_pointer_array_int
func invalidreturn_pointer_array_int() *[8]int

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

// Exported methods keep the C calling convention, but interface invoke thunks
// use the Go internal convention. When a method's result or a parameter is a
// large aggregate the two conventions place values differently, so calling
// such a method through an interface is not supported and must be reported
// instead of silently miscompiled.

type exportedBigResult struct{}

// ERROR: exported method (main.exportedBigResult).makeBig with a large aggregate result cannot be called through an interface
//
//export makeBigExported
func (exportedBigResult) makeBig() [1025]byte {
	return [1025]byte{}
}

type exportedBigParam struct{}

// ERROR: exported method (main.exportedBigParam).takeBig with a large aggregate parameter cannot be called through an interface
//
//export takeBigExported
func (exportedBigParam) takeBig(a [17]int32) int32 {
	return a[16]
}

func useExportedMethodsThroughInterfaces() {
	var m interface{ makeBig() [1025]byte } = exportedBigResult{}
	m.makeBig()
	var t interface{ takeBig([17]int32) int32 } = exportedBigParam{}
	t.takeBig([17]int32{})
}
