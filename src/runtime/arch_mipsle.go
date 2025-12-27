package runtime

const GOARCH = "mipsle"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

const callInstSize = 8 // "jal someFunc" is 4 bytes, plus a MIPS delay slot

const (
	linux_MAP_ANONYMOUS = 0x800
	linux_SIGBUS        = 10
	linux_SIGILL        = 4
	linux_SIGSEGV       = 11
)

// maxAlign is the maximum alignment required from the memory allocator.
// The o32 ABI requires 8-byte alignment for the stack and float64.
const maxAlign = 8

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}
