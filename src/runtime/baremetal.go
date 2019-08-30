// +build baremetal

package runtime

import (
	"unsafe"
)

//go:export _heap_start
var heapStartSymbol unsafe.Pointer

//go:export _heap_end
var heapEndSymbol unsafe.Pointer

//go:export _globals_start
var globalsStartSymbol unsafe.Pointer

//go:export _globals_end
var globalsEndSymbol unsafe.Pointer

//go:export _stack_top
var stackTopSymbol unsafe.Pointer

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(unsafe.Pointer(&heapEndSymbol))
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&globalsEndSymbol))
	stackTop     = uintptr(unsafe.Pointer(&stackTopSymbol))
)
