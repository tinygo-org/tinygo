//go:build (scheduler.tasks || scheduler.cores) && esp32p4

package task

import "unsafe"

// Returns a pointer where the system stack can be stored.
//
//go:linkname runtime_systemStackPtr runtime.systemStackPtr
func runtime_systemStackPtr() *uintptr

// calleeSavedRegs is the set of registers saved/restored during a goroutine
// context switch on ESP32-P4 (RV32IMAFC, ilp32f ABI).
//
// Layout must exactly match tinygo_swapTask in task_stack_esp32p4.S:
//
//	sp+ 0 .. sp+48 : s0–s11, ra  (integer callee-saved + return address)
//	sp+52          : fcsr        (FPU control/status register)
//	sp+56 .. sp+100: fs0–fs11   (FPU callee-saved registers)
//
// Total frame: 104 bytes.
type calleeSavedRegs struct {
	// Integer callee-saved registers
	s0  uintptr // x8/fp
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
	pc  uintptr // return address (ra)

	// FPU callee-saved registers (ilp32f ABI requires saving fs0–fs11)
	fcsr uint32 // FPU control/status (rounding mode + exception flags)
	fs0  uint32 // f8
	fs1  uint32 // f9
	fs2  uint32 // f18
	fs3  uint32 // f19
	fs4  uint32 // f20
	fs5  uint32 // f21
	fs6  uint32 // f22
	fs7  uint32 // f23
	fs8  uint32 // f24
	fs9  uint32 // f25
	fs10 uint32 // f26
	fs11 uint32 // f27
}

// archInit sets up the initial saved-register frame for a new goroutine.
// The goroutine stack is heap-allocated and zero-initialised, so FPU fields
// (fcsr=0 = round-to-nearest, no exception flags; fs0-fs11=0) are valid.
func (s *state) archInit(r *calleeSavedRegs, fn uintptr, args unsafe.Pointer) {
	// s.sp points to the bottom of the saved frame.  tinygo_swapTask will load
	// the integer registers from sp+0..sp+48 and the FPU regs from sp+52..sp+100.
	s.sp = uintptr(unsafe.Pointer(r))

	// tinygo_startTask reads s0 (function) and s1 (args pointer), then calls s0(s1).
	r.pc = uintptr(unsafe.Pointer(&startTask))
	r.s0 = fn
	r.s1 = uintptr(args)
}

func (s *state) resume() {
	swapTask(s.sp, runtime_systemStackPtr())
}

func (s *state) pause() {
	systemStackPtr := runtime_systemStackPtr()
	newStack := *systemStackPtr
	*systemStackPtr = 0
	swapTask(newStack, &s.sp)
}

// SystemStack returns the system stack pointer when called from a task stack.
// When called from the system stack, it returns 0.
func SystemStack() uintptr {
	return *runtime_systemStackPtr()
}

// OnSystemStack returns whether the caller is running on the system stack.
func OnSystemStack() bool {
	return SystemStack() == 0
}

//export tinygo_task_current
func tinygo_task_current() unsafe.Pointer {
	return unsafe.Pointer(Current())
}
