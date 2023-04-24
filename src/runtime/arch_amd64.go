package runtime

const GOARCH = "amd64"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 64

const deferExtraRegs = 0

const callInstSize = 5 // "call someFunction" is 5 bytes

// Align a pointer.
// Note that some amd64 instructions (like movaps) expect 16-byte aligned
// memory, thus the result must be 16-byte aligned.
func align(ptr uintptr) uintptr {
	return (ptr + 15) &^ 15
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
