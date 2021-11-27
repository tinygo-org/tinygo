package runtime

const GOARCH = "386"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 15) &^ 15
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
