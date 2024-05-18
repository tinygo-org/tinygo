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

// Return the current stack pointer using the llvm.stacksave.p0 intrinsic.
// It is normally used together with llvm.stackrestore.p0 but also works to get
// the current stack pointer in a platform-independent way.
func stacksave() unsafe.Pointer

//export strlen
func strlen(ptr unsafe.Pointer) uintptr

//export malloc
func malloc(size uintptr) unsafe.Pointer

// Compare two same-size buffers for equality.
func memequal(x, y unsafe.Pointer, n uintptr) bool {
	for i := uintptr(0); i < n; i++ {
		cx := *(*uint8)(unsafe.Add(x, i))
		cy := *(*uint8)(unsafe.Add(y, i))
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

// KeepAlive makes sure the value in the interface is alive until at least the
// point of the call.
func KeepAlive(x interface{})

var godebugUpdate func(string, string)

//go:linkname godebug_setUpdate internal/godebug.setUpdate
func godebug_setUpdate(update func(string, string)) {
	// The 'update' function needs to be called whenever the GODEBUG environment
	// variable changes (for example, via os.Setenv).
	godebugUpdate = update
}

//go:linkname godebug_setNewIncNonDefault internal/godebug.setNewIncNonDefault
func godebug_setNewIncNonDefault(newIncNonDefault func(string) func()) {
	// Dummy function necessary in Go 1.21.
}

// Write to the given file descriptor.
// This is called from internal/godebug starting with Go 1.21, and only seems to
// be called with the stderr file descriptor.
func write(fd uintptr, p unsafe.Pointer, n int32) int32 {
	if fd == 2 { // stderr
		// Convert to a string, because we know that p won't change during the
		// call to printstring.
		// TODO: use unsafe.String instead once we require Go 1.20.
		s := _string{
			ptr:    (*byte)(p),
			length: uintptr(n),
		}
		str := *(*string)(unsafe.Pointer(&s))
		printstring(str)
		return n
	}
	return 0
}
