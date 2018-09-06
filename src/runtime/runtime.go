package runtime

import (
	"unsafe"
)

const Compiler = "tgo"

// The compiler will fill this with calls to the initialization function of each
// package.
func initAll()

// These signatures are used to call the correct main function: with scheduling
// or without scheduling.
func main_main()
func main_mainAsync(parent *coroutine) *coroutine

// The compiler will change this to true if there are 'go' statements in the
// compiled program and turn it into a const.
var hasScheduler bool

// Entry point for Go. Initialize all packages and call main.main().
//go:export main
func main() int {
	// Run initializers of all packages.
	initAll()

	// This branch must be optimized away. Only one of the targets must remain,
	// or there will be link errors.
	if hasScheduler {
		// Initialize main and run the scheduler.
		coro := main_mainAsync(nil)
		scheduler(coro)
		return 0
	} else {
		// No scheduler is necessary. Call main directly.
		main_main()
		return 0
	}
}

func Sleep(d Duration) {
	// This function is treated specially by the compiler: when goroutines are
	// used, it is transformed into a llvm.coro.suspend() call.
	// When goroutines are not used this function behaves as normal.
	sleep(d)
}

func GOMAXPROCS(n int) int {
	// Note: setting GOMAXPROCS is ignored.
	return 1
}

func GOROOT() string {
	// TODO: don't hardcode but take the one at compile time.
	return "/usr/local/go"
}

// Copy size bytes from src to dst. The memory areas must not overlap.
func memcpy(dst, src unsafe.Pointer, size uintptr) {
	for i := uintptr(0); i < size; i++ {
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
