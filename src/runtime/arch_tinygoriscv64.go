//go:build tinygo.riscv64

package runtime

const GOARCH = "arm64" // riscv pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}
