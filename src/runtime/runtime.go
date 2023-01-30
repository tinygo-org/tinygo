package runtime

import (
	"unsafe"
)

//go:generate go run ../../tools/gen-critical-atomics -out ./atomics_critical.go

const Compiler = "tinygo"

// The compiler will fill this with calls to the initialization function of each
// package.
func initAll()

//go:linkname callMain main.main
func callMain()

func GOMAXPROCS(n int) int {
	// Note: setting GOMAXPROCS is ignored.
	return 1
}

func GOROOT() string {
	// TODO: don't hardcode but take the one at compile time.
	return "/usr/local/go"
}

// Copy size bytes from src to dst. The memory areas must not overlap.
// This function is implemented by the compiler as a call to a LLVM intrinsic
// like llvm.memcpy.p0.p0.i32(dst, src, size, false).
func memcpy(dst, src unsafe.Pointer, size uintptr)

// Copy size bytes from src to dst. The memory areas may overlap and will do the
// correct thing.
// This function is implemented by the compiler as a call to a LLVM intrinsic
// like llvm.memmove.p0.p0.i32(dst, src, size, false).
func memmove(dst, src unsafe.Pointer, size uintptr)

// Set the given number of bytes to zero.
// This function is implemented by the compiler as a call to a LLVM intrinsic
// like llvm.memset.p0.i32(ptr, 0, size, false).
func memzero(ptr unsafe.Pointer, size uintptr)

// This intrinsic returns the current stack pointer.
// It is normally used together with llvm.stackrestore but also works to get the
// current stack pointer in a platform-independent way.
//
//export llvm.stacksave
func stacksave() unsafe.Pointer

//export strlen
func strlen(ptr unsafe.Pointer) uintptr

//export malloc
func malloc(size uintptr) unsafe.Pointer

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
	return ticksToNanoseconds(ticks())
}

// Copied from the Go runtime source code.
//
//go:linkname os_sigpipe os.sigpipe
func os_sigpipe() {
	runtimePanic("too many writes on closed pipe")
}

// LockOSThread wires the calling goroutine to its current operating system thread.
// Stub for now
// Called by go1.18 standard library on windows, see https://github.com/golang/go/issues/49320
func LockOSThread() {
}

// UnlockOSThread undoes an earlier call to LockOSThread.
// Stub for now
func UnlockOSThread() {
}

// Mark KeepAlive as noinline so that it is easily detectable as an intrinsic.
//
//go:noinline

// KeepAlive marks its argument as currently reachable.
// This ensures that the object is not freed, and its finalizer is not run,
// before the point in the program where KeepAlive is called.
//
// A very simplified example showing where KeepAlive is required:
//
//	type File struct { d int }
//	d, err := syscall.Open("/file/path", syscall.O_RDONLY, 0)
//	// ... do something if err != nil ...
//	p := &File{d}
//	runtime.SetFinalizer(p, func(p *File) { syscall.Close(p.d) })
//	var buf [10]byte
//	n, err := syscall.Read(p.d, buf[:])
//	// Ensure p is not finalized until Read returns.
//	runtime.KeepAlive(p)
//	// No more uses of p after this point.
//
// Without the KeepAlive call, the finalizer could run at the start of
// syscall.Read, closing the file descriptor before syscall.Read makes
// the actual system call.
//
// Note: KeepAlive should only be used to prevent finalizers from
// running prematurely. In particular, when used with unsafe.Pointer,
// the rules for valid uses of unsafe.Pointer still apply.
func KeepAlive(x interface{}) {
	if alwaysFalse {
		println(x)
	}
}

// alwaysFalse is a boolean value that is always false.
// The compiler cannot see that alwaysFalse is always false,
// so it emits the test and keeps the call, giving the desired
// escape analysis result. The test is cheaper than the call.
var alwaysFalse bool

//go:linkname godebug_setUpdate internal/godebug.setUpdate
func godebug_setUpdate(update func(string, string)) {
	// Unimplemented. The 'update' function needs to be called whenever the
	// GODEBUG environment variable changes (for example, via os.Setenv).
}
