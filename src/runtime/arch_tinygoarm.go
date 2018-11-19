// +build tinygo.arm

package runtime

import (
	"unsafe"

	"device/arm"
)

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern _heap_start
var heapStartSymbol unsafe.Pointer

//go:extern _heap_end
var heapEndSymbol unsafe.Pointer

//go:extern _globals_start
var globalsStartSymbol unsafe.Pointer

//go:extern _globals_end
var globalsEndSymbol unsafe.Pointer

//go:extern _stack_top
var stackTopSymbol unsafe.Pointer

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(unsafe.Pointer(&heapEndSymbol))
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&globalsEndSymbol))
	stackTop     = uintptr(unsafe.Pointer(&stackTopSymbol))
)

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return arm.ReadRegister("sp")
}
