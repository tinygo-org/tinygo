// +build tinygo.riscv

package runtime

import "device/riscv"

const GOARCH = "arm" // riscv pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return riscv.AsmFull("mv {}, sp", nil)
}
