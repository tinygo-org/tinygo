// +build avr

package runtime

const GOARCH = "arm" // avr pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 8

// Align on a word boundary.
func align(ptr uintptr) uintptr {
	// No alignment necessary on the AVR.
	return ptr
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
