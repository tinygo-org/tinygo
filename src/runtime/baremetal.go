// +build baremetal

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

// FaultInfo contains details of the crash when it happens. All fields may be 0
// if the data is unavailable.
type FaultInfo struct {
	PC   uintptr // program counter of crashing instruction
	SP   uintptr // stack pointer at the time of the crash
	Code int     // optional system-dependent code (0 means "unknown" or "unimplemented")
}

var faultHandler func(info FaultInfo)

// SetFaultHandler registers a fault handler that is called when a critical
// problem occurs such as an illegal instruction or invalid memory read/write.
// It is designed to be set by the main package at startup, to handle critical
// problems in an application dependent way. For example, it may turn down a
// system in a safe way.
//
// The handler may be called in interrupt context at an unpredictable time
// (possibly even with interrupts disabled), therefore you shouldn't rely on any
// global state to be consistent. In particular, don't make heap allocations.
func SetFaultHandler(handler func(info FaultInfo)) {
	faultHandler = handler
}
