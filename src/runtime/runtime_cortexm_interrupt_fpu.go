//go:build cortexm && fpu

package runtime

// Grant full access to the FPU during the call to main()
const fpuEnabled = true

// The stack layout at the moment an interrupt occurs.
// Registers can be accessed if the stack pointer is cast to a pointer to this
// struct.
type interruptStack struct {
	R0    uintptr
	R1    uintptr
	R2    uintptr
	R3    uintptr
	R12   uintptr
	LR    uintptr
	PC    uintptr
	PSR   uintptr
	S0    uintptr
	S1    uintptr
	S2    uintptr
	S3    uintptr
	S4    uintptr
	S5    uintptr
	S6    uintptr
	S7    uintptr
	S8    uintptr
	S9    uintptr
	S10   uintptr
	S11   uintptr
	S12   uintptr
	S13   uintptr
	S14   uintptr
	S15   uintptr
	FPSCR uintptr
}
