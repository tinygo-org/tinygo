// +build scheduler.tasks,avr

package task

import "unsafe"

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_avr.S that relies on the
// exact layout of this struct.
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

// registers gets a pointer to the registers stored at the top of the stack.
func (s *state) registers() *calleeSavedRegs {
	return (*calleeSavedRegs)(unsafe.Pointer(s.sp + 1))
}

// startTask is a small wrapper function that sets up the first (and only)
// argument to the new goroutine and makes sure it is exited when the goroutine
// finishes.
//go:extern tinygo_startTask
var startTask [0]uint8

// archInit runs architecture-specific setup for the goroutine startup.
// Note: adding //go:noinline to work around an AVR backend bug.
//go:noinline
func (s *state) archInit(stack []uintptr, fn uintptr, args unsafe.Pointer) {
	// Set up the stack canary, a random number that should be checked when
	// switching from the task back to the scheduler. The stack canary pointer
	// points to the first word of the stack. If it has changed between now and
	// the next stack switch, there was a stack overflow.
	s.canaryPtr = &stack[0]
	*s.canaryPtr = stackCanary

	// Store the initial sp for the startTask function (implemented in assembly).
	s.sp = uintptr(unsafe.Pointer(&stack[uintptr(len(stack))-(unsafe.Sizeof(calleeSavedRegs{})/unsafe.Sizeof(uintptr(0)))])) - 1

	// Initialize the registers.
	// These will be popped off of the stack on the first resume of the goroutine.
	r := s.registers()

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
	switchToTask(s.sp)
}

//export tinygo_switchToTask
func switchToTask(uintptr)

//export tinygo_switchToScheduler
func switchToScheduler(*uintptr)

func (s *state) pause() {
	switchToScheduler(&s.sp)
}

//export tinygo_pause
func pause() {
	Pause()
}
