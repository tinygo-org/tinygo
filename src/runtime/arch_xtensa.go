//go:build xtensa
// +build xtensa

package runtime

const GOARCH = "arm" // xtensa pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

// Align on a word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
