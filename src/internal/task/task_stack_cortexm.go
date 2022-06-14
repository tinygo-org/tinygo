//go:build scheduler.tasks && cortexm
// +build scheduler.tasks,cortexm

package task

// Note that this is almost the same as task_stack_arm.go, but it uses the MSP
// register to store the system stack pointer instead of a global variable. The
// big advantage of this is that interrupts always execute with MSP (and not
// PSP, which is used for goroutines) so that goroutines do not need extra stack
// space for interrupts.

import (
	"device/arm"
	"unsafe"
)

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_cortexm.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	r4  uintptr
	r5  uintptr
	r6  uintptr
	r7  uintptr
	r8  uintptr
	r9  uintptr
	r10 uintptr
	r11 uintptr

	pc uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(r))

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.

	// Start the function at tinygo_startTask (defined in src/internal/task/task_stack_cortexm.S).
	// This assembly code calls a function (passed in r4) with a single argument (passed in r5).
	// After the function returns, it calls Pause().
	r.pc = uintptr(unsafe.Pointer(&startTask))

	// Pass the function to call in r4.
	// This function is a compiler-generated wrapper which loads arguments out of a struct pointer.
	// See createGoroutineStartWrapper (defined in compiler/goroutine.go) for more information.
	r.r4 = fn

	// Pass the pointer to the arguments struct in r5.
	r.r5 = uintptr(args)
}

func (s *state) switchTo(current *Task) {
	// This code is probably more complex than it needs to be.
	// TODO: simplify this, ideally to a single function call.
	if s == &mainTask.state {
		switchToMain(&current.state.sp)
	} else if current == &mainTask {
		switchToTask(s.sp)
	} else {
		swapTask(s.sp, &current.state.sp)
	}
}

//export tinygo_switchToTask
func switchToTask(uintptr)

//export tinygo_switchToMain
func switchToMain(*uintptr)

// SystemStack returns the system stack pointer. On Cortex-M, it is always
// available.
func SystemStack() uintptr {
	return arm.AsmFull("mrs {}, MSP", nil)
}
