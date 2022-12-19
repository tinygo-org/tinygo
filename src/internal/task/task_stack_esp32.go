//go:build scheduler.tasks && esp32

package task

// The windowed ABI (used on the ESP32) is as follows:
//   a0:    return address (link register)
//   a1:    stack pointer (must be 16-byte aligned)
//   a2-a7: incoming arguments
//   a7:    stack frame pointer (optional, normally unused in TinyGo)
// Sources:
//   http://cholla.mmto.org/esp8266/xtensa.html
//   https://0x04.net/~mwk/doc/xtensa.pdf

import (
	"unsafe"
)

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_esp8266.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	// Registers in the register window of tinygo_startTask.
	a0 uintptr
	a1 uintptr
	a2 uintptr
	a3 uintptr

	// Locals that can be used by tinygo_swapTask.
	// The first field is the a0 loaded in tinygo_swapTask, the rest is unused.
	locals [4]uintptr
}

// archInit runs architecture-specific setup for the goroutine startup.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// Store the stack pointer for the tinygo_swapTask function (implemented in
	// assembly). It needs to point to the locals field instead of a0 so that
	// the retw.n at the end of tinygo_swapTask will return into
	// tinygo_startTask with a0-a3 loaded (using the register window mechanism).
	s.sp = uintptr(unsafe.Pointer(&r.locals[0]))

	// Start the goroutine at tinygo_startTask (defined in
	// src/internal/task/task_stack_esp32.S). The topmost two bits are not part
	// of the address but instead store the register window of the caller.
	// In this case there is no caller, instead we set up the return address as
	// if tinygo_startTask called tinygo_swapTask with a call4 instruction.
	r.locals[0] = uintptr(unsafe.Pointer(&startTask))&^(3<<30) | (1 << 30)

	// Set up the stack pointer inside tinygo_startTask.
	// Unlike most calling conventions, the windowed ABI actually saves the
	// stack pointer on the stack to make register windowing work.
	r.a1 = uintptr(unsafe.Pointer(r)) + 32

	// Store the function pointer and the (only) parameter on the stack in a
	// location that will be reloaded into registers when doing the
	// pseudo-return to tinygo_startTask using the register window mechanism.
	r.a3 = fn
	r.a2 = uintptr(args)
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
