// +build xtensa

package runtime

import "device"

const GOARCH = "arm" // xtensa pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Align on a word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	// The stack pointer (sp) is a1.
	return device.AsmFull("mov {}, sp", nil)
}
