//go:build tinygo.riscv && virt && qemu
// +build tinygo.riscv,virt,qemu

package runtime

import (
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/device/riscv"
)

// This file implements the VirtIO RISC-V interface implemented in QEMU, which
// is an interface designed for emulation.

type timeUnit int64

var timestamp timeUnit

//export main
func main() {
	preinit()
	run()
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
}

// Memory-mapped I/O as defined by QEMU.
// Source: https://github.com/qemu/qemu/blob/master/hw/riscv/virt.c
// Technically this is an implementation detail but hopefully they won't change
// the memory-mapped I/O registers.
var (
	// UART0 output register.
	stdoutWrite = (*volatile.Register8)(unsafe.Pointer(uintptr(0x10000000)))
	// SiFive test finisher
	testFinisher = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))
)

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func abort() {
	exit(1)
}

func exit(code int) {
	// Make sure the QEMU process exits.
	if code == 0 {
		testFinisher.Set(0x5555) // FINISHER_PASS
	} else {
		// Exit code is stored in the upper 16 bits of the 32 bit value.
		testFinisher.Set(uint32(code)<<16 | 0x3333) // FINISHER_FAIL
	}

	// Lock up forever (as a fallback).
	for {
		riscv.Asm("wfi")
	}
}
