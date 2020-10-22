// +build cortexm3 cortexm4

package runtime

import (
	"device/arm"
	"unsafe"
)

const (
	hardFaultDebug = true
)

// See runtime_cortexm_hardfault_m0.go
//go:export handleHardFault
func handleHardFault(sp, msp *interruptStack, excReturn uintptr) {
	if excReturn&(1<<2) == 0 {
		sp = msp
	}

	fault := arm.GetFaultStatus()

	print("fatal error: ")
	if !hardFaultDebug {
		if uintptr(unsafe.Pointer(sp)) < stackTop-stackSize {
			print("stack overflow")
		} else {
			print("HardFault")
		}
		print(" with sp=", sp)

	} else {
		if uintptr(unsafe.Pointer(sp)) < stackTop-stackSize {
			print("stack overflow ")
		}

		if fault.Mem().InstructionAccessViolation() {
			print("instruction access violation")
		}
		if fault.Mem().DataAccessViolation() {
			print("data access violation")
		}
		if fault.Mem().WhileUnstackingException() {
			print(" while unstacking exception")
		}
		if fault.Mem().WhileStackingException() {
			print(" while stacking exception")
		}
		if fault.Mem().WhileLazyStackingException() {
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
		if fault.Bus().WhileLazyStackingException() {
			print(" during floating-point lazy state preservation")
		}

		if fault.Usage().UndefinedInstruction() {
			print("undefined instruction")
		}
		if fault.Usage().InvalidState() {
			print("invalid state change")
		}
		if fault.Usage().IllegalExceptionReturn() {
			print("illegal load of EXC_RETURN to the PC")
		}
		if fault.Usage().CoprocessorAccessViolation() {
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

		if sp != msp {
			print(" with psp=", sp)
		} else {
			print(" with msp=", sp)
		}
	}

	if uintptr(unsafe.Pointer(&sp.PC)) >= 0x20000000 {
		// Only print the PC if it points into memory.
		// It may not point into memory during a stack overflow, so check that
		// first before accessing the stack.
		print(" pc=", sp.PC)
	}
	print(" cfsr=", uintptr(fault))

	println()
	abort()
}
