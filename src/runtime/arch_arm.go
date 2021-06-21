// +build arm,!baremetal,!tinygo.wasm arm,arm7tdmi

package runtime

import "device/arm"

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return arm.AsmFull("mov {}, sp", nil)
}
