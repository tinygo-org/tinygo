//go:build cortexm

package runtime

import (
	"device/arm"
)

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 4 // "bl someFunction" is 4 bytes

// maxAlign is the maximum alignment required from the memory allocator.
// EABI requires 8-byte alignment for the stack and 64-bit values.
const maxAlign = 8

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

// The safest thing to do here would just be to disable interrupts for
// procPin/procUnpin. Note that a global variable is safe in this case, as any
// access to procPinnedMask will happen with interrupts disabled.

var procPinnedMask uintptr

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	procPinnedMask = arm.DisableInterrupts()
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	arm.EnableInterrupts(procPinnedMask)
}
