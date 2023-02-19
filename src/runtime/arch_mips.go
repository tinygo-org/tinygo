package runtime

const GOARCH = "mips"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const flag_linux_MAP_ANONYMOUS = 0x800

// Align on the maximum alignment for this platform (double).
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
