package runtime

import (
	"unsafe"
)

const Compiler = "tinygo"

// The compiler will fill this with calls to the initialization function of each
// package.
func initAll()

// callMain is a placeholder for the program main function.
// All references to this are replaced with references to the program main function by the compiler.
func callMain()

func GOMAXPROCS(n int) int {
	// Note: setting GOMAXPROCS is ignored.
	return 1
}

func GOROOT() string {
	// TODO: don't hardcode but take the one at compile time.
	return "/usr/local/go"
}

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	return nil
}

// Copy size bytes from src to dst. The memory areas must not overlap.
// Calls to this function are converted to LLVM intrinsic calls such as
// llvm.memcpy.p0i8.p0i8.i32(dst, src, size, false).
func memcpy(dst, src unsafe.Pointer, size uintptr)

// Copy size bytes from src to dst. The memory areas may overlap and will do the
// correct thing.
// Calls to this function are converted to LLVM intrinsic calls such as
// llvm.memmove.p0i8.p0i8.i32(dst, src, size, false).
func memmove(dst, src unsafe.Pointer, size uintptr)

// Set the given number of bytes to zero.
func memzero(ptr unsafe.Pointer, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = 0
	}
}

// Compare two same-size buffers for equality.
func memequal(x, y unsafe.Pointer, n uintptr) bool {
	for i := uintptr(0); i < n; i++ {
		cx := *(*uint8)(unsafe.Pointer(uintptr(x) + i))
		cy := *(*uint8)(unsafe.Pointer(uintptr(y) + i))
		if cx != cy {
			return false
		}
	}
	return true
}

func nanotime() int64 {
	return int64(ticks()) * tickMicros
}

// timeOffset is how long the monotonic clock started after the Unix epoch. It
// should be a positive integer under normal operation or zero when it has not
// been set.
var timeOffset int64

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = nanotime()
	sec = (mono + timeOffset) / (1000 * 1000 * 1000)
	nsec = int32((mono + timeOffset) - sec*(1000*1000*1000))
	return
}

// AdjustTimeOffset adds the given offset to the built-in time offset. A
// positive value adds to the time (skipping some time), a negative value moves
// the clock into the past.
func AdjustTimeOffset(offset int64) {
	// TODO: do this atomically?
	timeOffset += offset
}

// Copied from the Go runtime source code.
//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	runtimePanic("too many writes on closed pipe")
}

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string {
	return nil
}
