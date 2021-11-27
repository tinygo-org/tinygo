package runtime

const GOARCH = "arm64"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
