//go:build esp.clic

// Package esp drives the CPU interrupt controller of the Espressif RISC-V
// chips. The interrupt matrix routes a peripheral source to a CPU interrupt
// line, and this package controls the lines.
package esp

import (
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

// NumInterrupts is the highest CPU interrupt line. Line 0 is not used.
const NumInterrupts = 47

// MatrixOffset is added to a CPU interrupt line to get the value for an
// interrupt matrix map register. The first 16 CLIC sources are internal.
// See ESP-IDF CLIC_EXT_INTR_NUM_OFFSET.
const MatrixOffset = 16

// Each source has 4 bytes: ip, ie, attr, ctl.
const (
	clicBase    = 0x20800000
	clicIntBase = clicBase + 0x1000
)

// The level is in ctl bits 7:5 and the low bits must be 1.
// See ESP-IDF NLBITS_TO_BYTE.
const (
	clicNLBITS      = 3
	clicMaxPriority = 1<<clicNLBITS - 1
	clicLowBits     = 1<<(8-clicNLBITS) - 1
)

func clicReg(n, offset uint32) *volatile.Register8 {
	return (*volatile.Register8)(unsafe.Pointer(uintptr(clicIntBase + (n+MatrixOffset)*4 + offset)))
}

// Init masks all lines and sets the threshold to 0, the same as ESP-IDF
// RVHAL_INTR_ENABLE_THRESH.
func Init() {
	for i := uint32(1); i <= NumInterrupts; i++ {
		SetPriority(i, 0)
	}
	thresh := (*volatile.Register32)(unsafe.Pointer(uintptr(clicBase + 0x8)))
	thresh.Set(uint32(clicLowBits) << 24)
	riscv.MIE.Set(0xffffffff)
}

// Enable sets line n to edge trigger with the given priority and enables it.
func Enable(n, priority uint32) {
	clicReg(n, 1).Set(1)
	// Machine mode, rising edge, not hardware vectored.
	clicReg(n, 2).Set(0xc2)
	SetPriority(n, priority)
	clicReg(n, 0).Set(0)
}

// Priority returns the priority of line n.
func Priority(n uint32) uint32 {
	return uint32(clicReg(n, 3).Get() >> (8 - clicNLBITS))
}

// SetPriority sets the priority of line n. Priority 0 masks the line.
func SetPriority(n, priority uint32) {
	if priority > clicMaxPriority {
		priority = clicMaxPriority
	}
	clicReg(n, 3).Set(uint8(priority<<(8-clicNLBITS) | clicLowBits))
}

// Ack clears the pending state of line n.
func Ack(n uint32) {
	clicReg(n, 0).Set(0)
}

// Line returns the CPU interrupt line in mcause.
func Line(mcause uintptr) (uint32, bool) {
	code := uint32(mcause & 0xfff)
	if code <= MatrixOffset || code > MatrixOffset+NumInterrupts {
		return 0, false
	}
	return code - MatrixOffset, true
}
