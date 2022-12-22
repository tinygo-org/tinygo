//go:build scheduler.tasks && amd64 && windows

package task

// This is almost the same as task_stack_amd64.go, but with the extra rdi and
// rsi registers saved: Windows has a slightly different calling convention.

import "unsafe"

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_amd64_windows.S that relies on
// the exact layout of this struct.
// The calling convention is described here:
// https://docs.microsoft.com/en-us/cpp/build/x64-calling-convention?view=msvc-170
// Most importantly, these are the registers we need to save/restore:
//
// > The x64 ABI considers registers RBX, RBP, RDI, RSI, RSP, R12, R13, R14,
// > R15, and XMM6-XMM15 nonvolatile. They must be saved and restored by a
// > function that uses them.
type calleeSavedRegs struct {
	// Note: rbx is placed here so that the stack is correctly aligned when
	// loading/storing the xmm registers.
	rbx   uintptr
	xmm15 [2]uint64
	xmm14 [2]uint64
	xmm13 [2]uint64
	xmm12 [2]uint64
	xmm11 [2]uint64
	xmm10 [2]uint64
	xmm9  [2]uint64
	xmm8  [2]uint64
	xmm7  [2]uint64
	xmm6  [2]uint64
	rbp   uintptr
	rdi   uintptr
	rsi   uintptr
	r12   uintptr
	r13   uintptr
	r14   uintptr
	r15   uintptr

	pc uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in
	// src/internal/task/task_stack_amd64_windows.S). This assembly code calls a
	// function (passed in r12) with a single argument (passed in r13). After
	// the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in r12.
	// This function is a compiler-generated wrapper which loads arguments out
	// of a struct pointer. See createGoroutineStartWrapper (defined in
	// compiler/goroutine.go) for more information.
	r.r12 = fn

	// Pass the pointer to the arguments struct in r13.
	r.r13 = uintptr(args)
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
