package main

import "unsafe"

//go:wasmimport modulename voidimport
func voidimport(a int32, b float64)

//go:wasmimport modulename intimport
func intimport(p unsafe.Pointer, n uint32) uint32

var (
	voidValue = voidimport
	intValue  = intimport
)

func main() {
	voidValue(1, 2)
	println(intValue(nil, 0))
}
