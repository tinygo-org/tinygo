// +build scheduler.tasks,tinygo.riscv,tinygo.riscv.fp

package task

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_riscv.S that relies on the
// exact layout of this struct.
type calleeSavedRegs struct {
	fs0  uintptr // f8
	fs1  uintptr // f9
	fs2  uintptr // f18
	fs3  uintptr // f19
	fs4  uintptr // f20
	fs5  uintptr // f21
	fs6  uintptr // f22
	fs7  uintptr // f23
	fs8  uintptr // f24
	fs9  uintptr // f25
	fs10 uintptr // f26
	fs11 uintptr // f27

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
