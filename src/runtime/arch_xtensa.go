//go:build xtensa

package runtime

const GOARCH = "arm" // xtensa pretends to be arm

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 3 // "callx0 someFunction" (and similar) is 3 bytes

// maxAlign is the maximum alignment required from the memory allocator.
// The ABI requires 8-byte alignment for the stack and 64-bit types.
const maxAlign = 8

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
