//go:build scheduler.tasks && esp8266
// +build scheduler.tasks,esp8266

package task

// Stack switch implementation for the ESP8266, which does not use the windowed
// ABI of Xtensa. Registers are assigned as follows:
//   a0:      return address (link register)
//   a1:      stack pointer (must be 16-byte aligned)
//   a2-a7:   incoming arguments
//   a8:      static chain (unused)
//   a12-a15: callee-saved
//   a15:     stack frame pointer (optional, unused)
// Sources:
// http://cholla.mmto.org/esp8266/xtensa.html
// https://0x04.net/~mwk/doc/xtensa.pdf

import "unsafe"

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_esp8266.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	a12 uintptr
	a13 uintptr
	a14 uintptr
	a15 uintptr

	pc uintptr // also link register or r0
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in
	// src/internal/task/task_stack_esp8266.S).
	// This assembly code calls a function (passed in a12) with a single argument
	// (passed in a13). After the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in a12.
	// This function is a compiler-generated wrapper which loads arguments out of a struct pointer.
	// See createGoroutineStartWrapper (defined in compiler/goroutine.go) for more information.
	r.a12 = fn

	// Pass the pointer to the arguments struct in a13.
	r.a13 = uintptr(args)
}

func (s *state) resume() {
	swapTask(s.sp, &systemStack)
}

func (s *state) pause() {
	newStack := systemStack
	systemStack = 0
	swapTask(newStack, &s.sp)
}

// SystemStack returns the system stack pointer when called from a task stack.
// When called from the system stack, it returns 0.
func SystemStack() uintptr {
	return systemStack
}
