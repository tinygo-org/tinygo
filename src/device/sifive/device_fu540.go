//go:build fu540

// This file implements things not properly implemented in the svd for fu540,
// as they were in the fe310.
package sifive

import (
	"runtime/volatile"
	"unsafe"
)

// Coreplex Local Interrupts
type CLINT_Type struct {
	MSIP     [5]volatile.Register32 // 0x0
	_        [0x4000 - 5*32]byte
	MTIMECMP [5]volatile.Register64 // 0x4000
	_        [0xbff8 - (5*64 - 0x4000 - 5*32)]byte
	MTIME    volatile.Register64 // 0xBFF8
}

var (
	// Coreplex Local Interrupts
	CLINT = (*CLINT_Type)(unsafe.Pointer(uintptr(0x2000000)))
)
