//go:build scheduler.tasks && amd64 && uefi

package task

// This is almost the same as task_stack_amd64.go, but with the extra rdi and
// rsi registers saved: UEFI on amd64 uses the Win64 ABI.

import "unsafe"

var systemStack uintptr

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see task_stack_amd64_windows.S that relies on
// the exact layout of this struct.
type calleeSavedRegs struct {
	// rbx is placed here so the stack is correctly aligned when saving XMM regs.
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

func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	s.sp = uintptr(unsafe.Pointer(r))
	r.pc = uintptr(unsafe.Pointer(&startTask))
	r.r12 = fn
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

func SystemStack() uintptr {
	return systemStack
}
