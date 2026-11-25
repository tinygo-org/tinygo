package xtensa

import "C"

// DisableInterrupts disables all interrupts, and returns the old interrupt
// state.
//
//export DisableInterrupts
func DisableInterrupts() uintptr

// EnableInterrupts enables all interrupts again. The value passed in must be
// the mask returned by DisableInterrupts.
//
//export EnableInterrupts
func EnableInterrupts(mask uintptr)

// ReadINTENABLE reads the INTENABLE special register.
//
//export ReadINTENABLE
func ReadINTENABLE() uint32

// WriteINTENABLE writes the INTENABLE special register.
//
//export WriteINTENABLE
func WriteINTENABLE(val uint32)

// ReadINTERRUPT reads the INTERRUPT special register, which holds the pending
// CPU interrupts.
//
//export ReadINTERRUPT
func ReadINTERRUPT() uint32

// WriteINTCLEAR writes the INTCLEAR special register.
//
//export WriteINTCLEAR
func WriteINTCLEAR(val uint32)

// WriteVECBASE writes the VECBASE special register, the address of the vector
// table.
//
//export WriteVECBASE
func WriteVECBASE(val uintptr)

// ReadPS reads the PS special register.
//
//export ReadPS
func ReadPS() uintptr
