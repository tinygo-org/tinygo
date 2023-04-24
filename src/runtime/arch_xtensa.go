//go:build xtensa

package runtime

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
