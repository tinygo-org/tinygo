package runtime

import (
	"unsafe"
)

const Compiler = "tgo"

// The compiler will fill this with calls to the initialization function of each
// package.
func initAll()

// A function call to this function is replaced withone of the following,
// depending on whether the scheduler is necessary:
//
// Without scheduler:
//
//     main.main()
//
// With scheduler:
//
//     main.main()
//     scheduler()
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
func memcpy(dst, src unsafe.Pointer, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(dst) + i)) = *(*uint8)(unsafe.Pointer(uintptr(src) + i))
	}
}

// Copy size bytes from src to dst. The memory areas may overlap and will do the
// correct thing.
func memmove(dst, src unsafe.Pointer, size uintptr) {
	if uintptr(dst) < uintptr(src) {
		// Copy forwards.
		memcpy(dst, src, size)
		return
	}
	// Copy backwards.
	for i := size; i != 0; {
		i--
		*(*uint8)(unsafe.Pointer(uintptr(dst) + i)) = *(*uint8)(unsafe.Pointer(uintptr(src) + i))
	}
}

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

//go:linkname sleep time.Sleep
func sleep(d int64) {
	sleepTicks(timeUnit(d / tickMicros))
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = int64(ticks()) * tickMicros
	sec = mono / (1000 * 1000 * 1000)
	nsec = int32(mono - sec*(1000*1000*1000))
	return
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
