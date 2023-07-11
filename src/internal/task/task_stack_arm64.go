//go:build scheduler.tasks && arm64

package task

import "unsafe"

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_arm64.S that relies on the exact
// layout of this struct.
type calleeSavedRegs struct {
	x19 uintptr
	x20 uintptr
	x21 uintptr
	x22 uintptr
	x23 uintptr
	x24 uintptr
	x25 uintptr
	x26 uintptr
	x27 uintptr
	x28 uintptr
	x29 uintptr
	pc  uintptr // aka x30 aka LR

	d8  uintptr
	d9  uintptr
	d10 uintptr
	d11 uintptr
	d12 uintptr
	d13 uintptr
	d14 uintptr
	d15 uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in src/internal/task/task_stack_arm64.S).
	// This assembly code calls a function (passed in x19) with a single argument
	// (passed in x20). After the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in x19.
	// This function is a compiler-generated wrapper which loads arguments out of a struct pointer.
	// See createGoroutineStartWrapper (defined in compiler/goroutine.go) for more information.
	r.x19 = fn

	// Pass the pointer to the arguments struct in x20.
	r.x20 = uintptr(args)
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
