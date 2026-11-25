//go:build xtensa

package runtime

import "device/xtensa"

const GOARCH = "arm" // xtensa pretends to be arm

const zeroSizeAllocPtr uintptr = 16 // part of early flash: partition table, etc

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Xtensa's windowed ABI uses a 17-word jmp_buf. JumpSP and JumpPC provide the
// first two words.
const deferExtraRegs = 15

const callInstSize = 3 // "callx0 someFunction" (and similar) is 3 bytes

// The largest alignment according to the Xtensa ABI is 8 (long long, double).
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

var procPinnedMask uintptr

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	procPinnedMask = xtensa.DisableInterrupts()
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	xtensa.EnableInterrupts(procPinnedMask)
}
