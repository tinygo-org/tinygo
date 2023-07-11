//go:build scheduler.tasks && avr

package task

import "unsafe"

//go:extern tinygo_systemStack
var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_avr.S that relies on the exact
// layout of this struct.
//
// https://gcc.gnu.org/wiki/avr-gcc#Call-Saved_Registers
type calleeSavedRegs struct {
	r2r3   uintptr
	r4r5   uintptr
	r6r7   uintptr
	r8r9   uintptr
	r10r11 uintptr
	r12r13 uintptr
	r14r15 uintptr
	r16r17 uintptr
	r28r29 uintptr // Y register

	pc uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
// Note: adding //go:noinline to work around an AVR backend bug.
//
//go:noinline
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r)) - 1

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask.
	startTask := uintptr(unsafe.Pointer(&startTask))
	r.pc = startTask/2>>8 | startTask/2<<8

	// Pass the function to call in r2:r3.
	// This function is a compiler-generated wrapper which loads arguments out
	// of a struct pointer. See createGoroutineStartWrapper (defined in
	// compiler/goroutine.go) for more information.
	r.r2r3 = fn

	// Pass the pointer to the arguments struct in r4:45.
	r.r4r5 = uintptr(args)
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
