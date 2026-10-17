package main

import "unsafe"

//go:align 256
var byteA byte

//go:align 256
var byteB byte

var (
	ptrA = &byteA
	ptrB = &byteB
)

// Interp runs this call at compile time. The bytes of a pointer are known only
// at runtime, so interp must not fold the comparison.
var initEqual = comparePointerBytes()

//go:noinline
func comparePointerBytes() bool {
	return *(*byte)(unsafe.Pointer(&ptrA)) == *(*byte)(unsafe.Pointer(&ptrB))
}

func main() {
	// The result of the comparison depends on the target.
	println("pointer bytes:", initEqual == comparePointerBytes())
}
