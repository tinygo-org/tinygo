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

// The following functions are workarounds for things missing in compiler-rt.
// They will likely need special assembly implementations.
// They are treated specially: they're added to @llvm.compiler.used so that the
// linker won't eliminate them.

//export __mulsi3
func __mulsi3(a, b uint32) uint32 {
	var r uint32
	for a != 0 {
		if a&1 != 0 {
			r += b
		}
		a >>= 1
		b <<= 1
	}
	return r
}

//export __divsi3
func __divsi3(a, b int32) int32

//export __udivsi3
func __udivsi3(a, b uint32) uint32

//export __divmodsi4
func __divmodsi4(a, b int32) uint64 {
	d := __divsi3(a, b)
	rem := a - (d * b)
	return uint64(uint32(d)) | uint64(uint32(rem))<<32
}

//export __udivmodsi4
func __udivmodsi4(a, b uint32) uint64 {
	d := __udivsi3(a, b)
	rem := a - (d * b)
	return uint64(d) | uint64(rem)<<32
}
