// +build avr cortexm tinygo.riscv

package runtime

import (
	"unsafe"
)

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
