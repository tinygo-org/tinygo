// +build fe310

package main

import (
	"device/riscv"
	"runtime/volatile"
	"unsafe"
)

func callTrap() {
	// Run an undefined instruction.
	// With the C extension, it decodes to:
	//     undef
	//     undef
	riscv.Asm(".long 0")
}

// Special memory-mapped device to exit tests in QEMU, created by SiFive.
var testExit = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

func abort() {
	// Signal a successful exit.
	testExit.Set(0x5555)
}
