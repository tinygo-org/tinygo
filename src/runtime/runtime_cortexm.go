// +build cortexm

package runtime

import (
	"device/arm"
	"unsafe"
)

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//go:extern _sdata
var _sdata [0]byte

//go:extern _sidata
var _sidata [0]byte

//go:extern _edata
var _edata [0]byte

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

// calleeSavedRegs is the list of registers that must be saved and restored when
// switching between tasks. Also see scheduler_cortexm.S that relies on the
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
}

// prepareStartTask stores fn and args in some callee-saved registers that can
// then be used by the startTask function (implemented in assembly) to set up
// the initial stack pointer and initial argument with the pointer to the object
// with the goroutine start arguments.
func (r *calleeSavedRegs) prepareStartTask(fn, args uintptr) {
	r.r4 = fn
	r.r5 = args
}

func abort() {
	// disable all interrupts
	arm.DisableInterrupts()

	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

// The stack layout at the moment an interrupt occurs.
// Registers can be accessed if the stack pointer is cast to a pointer to this
// struct.
type interruptStack struct {
	R0  uintptr
	R1  uintptr
	R2  uintptr
	R3  uintptr
	R12 uintptr
	LR  uintptr
	PC  uintptr
	PSR uintptr
}

// This function is called at HardFault.
// Before this function is called, the stack pointer is reset to the initial
// stack pointer (loaded from addres 0x0) and the previous stack pointer is
// passed as an argument to this function. This allows for easy inspection of
// the stack the moment a HardFault occurs, but it means that the stack will be
// corrupted by this function and thus this handler must not attempt to recover.
//
// For details, see:
// https://community.arm.com/developer/ip-products/system/f/embedded-forum/3257/debugging-a-cortex-m0-hard-fault
// https://blog.feabhas.com/2013/02/developing-a-generic-hard-fault-handler-for-arm-cortex-m3cortex-m4/
//go:export handleHardFault
func handleHardFault(sp *interruptStack) {
	fault := arm.GetFaultStatus()
	spValid := !fault.Bus().ImpreciseDataBusError()

	print("fatal error: ")
	if spValid && uintptr(unsafe.Pointer(sp)) < 0x20000000 {
		print("stack overflow")
	} else {
		if fault.Mem().InstructionAccessViolation() {
			print("instruction access violation")
		}
		if fault.Mem().DataAccessViolation() {
			print("data access violation")
		}
		if fault.Mem().WhileUnstackingException() {
			print(" while unstacking exception")
		}
		if fault.Mem().WileStackingException() {
			print(" while stacking exception")
		}
		if fault.Mem().DuringFPLazyStatePres() {
			print(" during floating-point lazy state preservation")
		}

		if fault.Bus().InstructionBusError() {
			print("instruction bus error")
		}
		if fault.Bus().PreciseDataBusError() {
			print("data bus error (precise)")
		}
		if fault.Bus().ImpreciseDataBusError() {
			print("data bus error (imprecise)")
		}
		if fault.Bus().WhileUnstackingException() {
			print(" while unstacking exception")
		}
		if fault.Bus().WhileStackingException() {
			print(" while stacking exception")
		}
		if fault.Bus().DuringFPLazyStatePres() {
			print(" during floating-point lazy state preservation")
		}

		if fault.Usage().UndefinedInstruction() {
			print("undefined instruction")
		}
		if fault.Usage().IllegalUseOfEPSR() {
			print("illegal use of the EPSR")
		}
		if fault.Usage().IllegalExceptionReturn() {
			print("illegal load of EXC_RETURN to the PC")
		}
		if fault.Usage().AttemptedToAccessCoprocessor() {
			print("coprocessor access violation")
		}
		if fault.Usage().UnalignedMemoryAccess() {
			print("unaligned memory access")
		}
		if fault.Usage().DivideByZero() {
			print("divide by zero")
		}

		if fault.Unknown() {
			print("unknown hard fault")
		}

		if addr, ok := fault.Mem().Address(); ok {
			print(" with fault address ", addr)
		}

		if addr, ok := fault.Bus().Address(); ok {
			print(" with bus fault address ", addr)
		}
	}
	if spValid {
		print(" with sp=", sp)
		if uintptr(unsafe.Pointer(&sp.PC)) >= 0x20000000 {
			// Only print the PC if it points into memory.
			// It may not point into memory during a stack overflow, so check that
			// first before accessing the stack.
			print(" pc=", sp.PC)
		}
	}
	println()
	abort()
}

// Implement memset for LLVM and compiler-rt.
//go:export memset
func libc_memset(ptr unsafe.Pointer, c byte, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
}

// Implement memmove for LLVM and compiler-rt.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) {
	memmove(dst, src, size)
}

// Implement memcpy for LLVM and compiler-rt.
//go:export memcpy
func libc_memcpy(dst, src unsafe.Pointer, size uintptr) {
	memcpy(dst, src, size)
}
