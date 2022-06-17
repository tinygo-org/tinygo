//go:build avr
// +build avr

package runtime

import "runtime/interrupt"

const GOARCH = "arm" // avr pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 8

const deferExtraRegs = 1 // the frame pointer (Y register) also needs to be stored

// Align on a word boundary.
func align(ptr uintptr) uintptr {
	// No alignment necessary on the AVR.
	return ptr
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

// The safest thing to do here would just be to disable interrupts for
// procPin/procUnpin. Note that a global variable is safe in this case, as any
// access to procPinnedMask will happen with interrupts disabled.

var procPinnedMask interrupt.State

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	procPinnedMask = interrupt.Disable()
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	interrupt.Restore(procPinnedMask)
}
