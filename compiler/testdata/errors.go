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

//go:wasmimport modulename validreturn1
func validreturn1() int32

//go:wasmimport modulename validreturn2
func validreturn2() int

//go:wasmimport modulename validreturn3
func validreturn3() *int32

//go:wasmimport modulename validreturn4
func validreturn4() *S

//go:wasmimport modulename validreturn5
func validreturn5() unsafe.Pointer

// ERROR: //go:wasmimport modulename manyreturns: too many return values
//
//go:wasmimport modulename manyreturns
func manyreturns() (int32, int32)

// ERROR: //go:wasmimport modulename invalidreturn1: unsupported result type func()
//
//go:wasmimport modulename invalidreturn1
func invalidreturn1() func()

// ERROR: //go:wasmimport modulename invalidreturn2: unsupported result type []byte
//
//go:wasmimport modulename invalidreturn2
func invalidreturn2() []byte

// ERROR: //go:wasmimport modulename invalidreturn3: unsupported result type chan int
//
//go:wasmimport modulename invalidreturn3
func invalidreturn3() chan int

// ERROR: //go:wasmimport modulename invalidreturn4: unsupported result type string
//
//go:wasmimport modulename invalidreturn4
func invalidreturn4() string
