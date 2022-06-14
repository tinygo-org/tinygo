//go:build scheduler.tasks && 386
// +build scheduler.tasks,386

package task

import "unsafe"

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_386.S that relies on the exact
// layout of this struct.
type calleeSavedRegs struct {
	ebx uintptr
	esi uintptr
	edi uintptr
	ebp uintptr

	pc uintptr

	// Pad this struct so that tasks start on a 16-byte aligned stack.
	_ [3]uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in
	// src/internal/task/task_stack_386.S). This assembly code calls a function
	// (passed in EBX) with a single argument (passed in ESI). After the
	// function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in EBX.
	// This function is a compiler-generated wrapper which loads arguments out
	// of a struct pointer. See createGoroutineStartWrapper (defined in
	// compiler/goroutine.go) for more information.
	r.ebx = fn

	// Pass the pointer to the arguments struct in ESI.
	r.esi = uintptr(args)
}

func (s *state) switchTo(current *Task) {
	swapTask(s.sp, &current.state.sp)
}

// SystemStack returns the system stack pointer when called from a task stack.
// When called from the system stack, it returns 0.
func SystemStack() uintptr {
	return mainTask.state.sp
}
