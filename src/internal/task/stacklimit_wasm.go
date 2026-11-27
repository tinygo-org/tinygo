//go:build tinygo.wasm

package task

import "unsafe"

//go:extern __stack_low
var stackLowSymbol [0]byte

// stackGuard is the stack space kept free for stackOverflow to print.
const stackGuard = 2048

// stackLimit is the lowest stack pointer allowed in a function with a stack
// frame, or 0 for no limit. The checks are added by transform.AddStackLimitChecks.
var stackLimit uintptr

// setStackLimit sets the limit for a stack that starts at base and returns
// the limit to restore afterwards.
func setStackLimit(base uintptr) uintptr {
	prev := stackLimit
	if prev == 0 {
		// A static initializer would run in initAll on the main goroutine.
		prev = uintptr(unsafe.Pointer(&stackLowSymbol)) + stackGuard
	}
	stackLimit = base + stackGuard
	return prev
}

func restoreStackLimit(limit uintptr) {
	stackLimit = limit
}

// stackOverflow is called when the stack pointer is below stackLimit.
//
//go:noinline
func stackOverflow() {
	stackLimit = 0
	runtimeFatal("stack overflow")
}
