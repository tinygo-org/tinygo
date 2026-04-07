//go:build xtensa

package runtime

import "device"

const GOARCH = "arm" // xtensa pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 3 // "callx0 someFunction" (and similar) is 3 bytes

// The largest alignment according to the Xtensa ABI is 8 (long long, double).
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

// Disable interrupts for procPin/procUnpin using the Xtensa RSIL/WSR PS
// instructions.  A global variable is safe here because accesses happen
// with interrupts disabled.
var procPinnedMask uintptr

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	// rsil sets PS.INTLEVEL=15 (mask all) and returns the old PS.
	procPinnedMask = uintptr(device.AsmFull("rsil {}, 15", nil))
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	device.AsmFull("wsr {state}, PS", map[string]interface{}{
		"state": procPinnedMask,
	})
}
