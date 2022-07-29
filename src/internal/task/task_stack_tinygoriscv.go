//go:build scheduler.tasks && tinygo.riscv
// +build scheduler.tasks,tinygo.riscv

package task

import "unsafe"

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_riscv.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	s0  uintptr // x8 (fp)
	s1  uintptr // x9
	s2  uintptr // x18
	s3  uintptr // x19
	s4  uintptr // x20
	s5  uintptr // x21
	s6  uintptr // x22
	s7  uintptr // x23
	s8  uintptr // x24
	s9  uintptr // x25
	s10 uintptr // x26
	s11 uintptr // x27

	pc uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in src/internal/task/task_stack_riscv.S).
	// This assembly code calls a function (passed in s0) with a single argument
	// (passed in s1). After the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in s0.
	// This function is a compiler-generated wrapper which loads arguments out
	// of a struct pointer. See createGoroutineStartWrapper (defined in
	// compiler/goroutine.go) for more information.
	r.s0 = fn

	// Pass the pointer to the arguments struct in s1.
	r.s1 = uintptr(args)
}

func (s *state) switchTo(current *Task) {
	swapTask(s.sp, &current.state.sp)
}

// SystemStack returns the system stack pointer when called from a task stack.
// When called from the system stack, it returns 0.
func SystemStack() uintptr {
	return mainTask.state.sp
}
