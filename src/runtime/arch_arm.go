//go:build (arm && !baremetal && !tinygo.wasm) || (arm && arm7tdmi)

package runtime

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 4 // "bl someFunction" is 4 bytes

const (
	linux_MAP_ANONYMOUS = 0x20
	linux_SIGBUS        = 7
	linux_SIGILL        = 4
	linux_SIGSEGV       = 11
)

// maxAlign is the maximum alignment required from the memory allocator.
// EABI requires 8-byte alignment for the stack and 64-bit values.
const maxAlign = 8

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
