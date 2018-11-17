// +build amd64

package runtime

const GOARCH = "amd64"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}
