package runtime

const GOARCH = "mips"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 8 // "jal someFunc" is 4 bytes, plus a MIPS delay slot

const linux_MAP_ANONYMOUS = 0x800

// It appears that MIPS has a maximum alignment of 8 bytes.
func align(ptr uintptr) uintptr {
	return (ptr + 7) &^ 7
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
