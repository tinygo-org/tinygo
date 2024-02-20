//go:build fu540

// This file implements things not properly implemented in the svd for fu540,
// as they were in the fe310.
package sifive

var (
	// Coreplex Local Interrupts
	CLINT = (*CLINT_Type)(unsafe.Pointer(uintptr(0x2000000)))
)
