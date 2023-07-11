//go:build atsamd21 || nrf51

package runtime

import (
	"unsafe"
)

// This function is called at HardFault.
// Before this function is called, the stack pointer is reset to the initial
// stack pointer (loaded from addres 0x0) and the previous stack pointer is
// passed as an argument to this function. This allows for easy inspection of
// the stack the moment a HardFault occurs, but it means that the stack will be
// corrupted by this function and thus this handler must not attempt to recover.
//
// For details, see:
// https://community.arm.com/developer/ip-products/system/f/embedded-forum/3257/debugging-a-cortex-m0-hard-fault
// https://blog.feabhas.com/2013/02/developing-a-generic-hard-fault-handler-for-arm-cortex-m3cortex-m4/
//
//export handleHardFault
func handleHardFault(sp *interruptStack) {
	print("fatal error: ")
	if uintptr(unsafe.Pointer(sp)) < 0x20000000 {
		print("stack overflow")
	} else {
		// TODO: try to find the cause of the hard fault. Especially on
		// Cortex-M3 and higher it is possible to find more detailed information
		// in special status registers.
		print("HardFault")
	}
	print(" with sp=", sp)
	if uintptr(unsafe.Pointer(&sp.PC)) >= 0x20000000 {
		// Only print the PC if it points into memory.
		// It may not point into memory during a stack overflow, so check that
		// first before accessing the stack.
		print(" pc=", sp.PC)
	}
	println()
	abort()
}
