//go:build (arm && !baremetal && !tinygo.wasm) || (arm && arm7tdmi)
// +build arm,!baremetal,!tinygo.wasm arm,arm7tdmi

package runtime

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
