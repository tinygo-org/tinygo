//go:build cortexm && !atsamd21 && !nrf51
// +build cortexm,!atsamd21,!nrf51

package runtime

import (
	"unsafe"

	"tinygo.org/x/device/arm"
)

const (
	SCB_CFSR_KnownFault = arm.SCB_CFSR_IACCVIOL | arm.SCB_CFSR_DACCVIOL |
		arm.SCB_CFSR_MUNSTKERR | arm.SCB_CFSR_MSTKERR | arm.SCB_CFSR_MLSPERR |
		arm.SCB_CFSR_IBUSERR | arm.SCB_CFSR_PRECISERR | arm.SCB_CFSR_IMPRECISERR |
		arm.SCB_CFSR_UNSTKERR | arm.SCB_CFSR_STKERR | arm.SCB_CFSR_LSPERR |
		arm.SCB_CFSR_UNDEFINSTR | arm.SCB_CFSR_INVSTATE | arm.SCB_CFSR_INVPC |
		arm.SCB_CFSR_NOCP | arm.SCB_CFSR_UNALIGNED | arm.SCB_CFSR_DIVBYZERO
)

// See runtime_cortexm_hardfault.go
//
//go:export handleHardFault
func handleHardFault(sp *interruptStack) {
	fault := GetFaultStatus()
	spValid := !fault.Bus().ImpreciseDataBusError()

	print("fatal error: ")
	if spValid && uintptr(unsafe.Pointer(sp)) < 0x20000000 {
		print("stack overflow? ")
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

// Descriptions are sourced from the K66 SVD and
// http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.dui0552a/Cihcfefj.html

// GetFaultStatus reads the System Control Block Configurable Fault Status
// Register and returns it as a FaultStatus.
func GetFaultStatus() FaultStatus {
	return FaultStatus(arm.SCB.CFSR.Get())
}

type FaultStatus uint32
type MemFaultStatus uint32
type BusFaultStatus uint32
type UsageFaultStatus uint32

func (fs FaultStatus) Mem() MemFaultStatus     { return MemFaultStatus(fs) }
func (fs FaultStatus) Bus() BusFaultStatus     { return BusFaultStatus(fs) }
func (fs FaultStatus) Usage() UsageFaultStatus { return UsageFaultStatus(fs) }

// Unknown returns true if the cause of the fault is not know
func (fs FaultStatus) Unknown() bool {
	return fs&SCB_CFSR_KnownFault == 0
}

// InstructionAccessViolation: the processor attempted an instruction fetch from
// a location that does not permit execution
//
// "This fault occurs on any access to an XN region, even when the MPU is
// disabled or not present.
//
// When this bit is 1, the PC value stacked for the exception return points to
// the faulting instruction. The processor has not written a fault address to
// the MMAR."
func (fs MemFaultStatus) InstructionAccessViolation() bool {
	return fs&arm.SCB_CFSR_IACCVIOL != 0
}

// DataAccessViolation: the processor attempted a load or store at a location
// that does not permit the operation
//
// "When this bit is 1, the PC value stacked for the exception return points to
// the faulting instruction. The processor has loaded the MMAR with the address
// of the attempted access."
func (fs MemFaultStatus) DataAccessViolation() bool {
	return fs&arm.SCB_CFSR_DACCVIOL != 0
}

// WhileUnstackingException: unstack for an exception return has caused one or
// more access violations
//
// "This fault is chained to the handler. This means that when this bit is 1,
// the original return stack is still present. The processor has not adjusted
// the SP from the failing return, and has not performed a new save. The
// processor has not written a fault address to the MMAR."
func (fs MemFaultStatus) WhileUnstackingException() bool {
	return fs&arm.SCB_CFSR_MUNSTKERR != 0
}

// WileStackingException: stacking for an exception entry has caused one or more
// access violations
//
// "When this bit is 1, the SP is still adjusted but the values in the context
// area on the stack might be incorrect. The processor has not written a fault
// address to the MMAR."
func (fs MemFaultStatus) WileStackingException() bool {
	return fs&arm.SCB_CFSR_MSTKERR != 0
}

// DuringFPLazyStatePres: A MemManage fault occurred during floating-point lazy
// state preservation
func (fs MemFaultStatus) DuringFPLazyStatePres() bool {
	return fs&arm.SCB_CFSR_MLSPERR != 0
}

// InstructionBusError: instruction bus error
//
// "The processor detects the instruction bus error on prefetching an
// instruction, but it sets the IBUSERR flag to 1 only if it attempts to issue
// the faulting instruction.
//
// When the processor sets this bit is 1, it does not write a fault address to
// the BFAR."
func (fs BusFaultStatus) InstructionBusError() bool {
	return fs&arm.SCB_CFSR_IBUSERR != 0
}

// PreciseDataBusError: a data bus error has occurred, and the PC value stacked
// for the exception return points to the instruction that caused the fault
func (fs BusFaultStatus) PreciseDataBusError() bool {
	return fs&arm.SCB_CFSR_PRECISERR != 0
}

// ImpreciseDataBusError: a data bus error has occurred, but the return address
// in the stack frame is not related to the instruction that caused the error
//
// "When the processor sets this bit to 1, it does not write a fault address to
// the BFAR.
//
// This is an asynchronous fault. Therefore, if it is detected when the priority
// of the current process is higher than the BusFault priority, the BusFault
// becomes pending and becomes active only when the processor returns from all
// higher priority processes. If a precise fault occurs before the processor
// enters the handler for the imprecise BusFault, the handler detects both
// IMPRECISERR set to 1 and one of the precise fault status bits set to 1."
func (fs BusFaultStatus) ImpreciseDataBusError() bool {
	return fs&arm.SCB_CFSR_IMPRECISERR != 0
}

// WhileUnstackingException: unstack for an exception return has caused one or
// more BusFaults
//
// "This fault is chained to the handler. This means that when the processor
// sets this bit to 1, the original return stack is still present. The processor
// does not adjust the SP from the failing return, does not performed a new
// save, and does not write a fault address to the BFAR."
func (fs BusFaultStatus) WhileUnstackingException() bool {
	return fs&arm.SCB_CFSR_UNSTKERR != 0
}

// WhileStackingException: stacking for an exception entry has caused one or
// more BusFaults
//
// "When the processor sets this bit to 1, the SP is still adjusted but the
// values in the context area on the stack might be incorrect. The processor
// does not write a fault address to the BFAR."
func (fs BusFaultStatus) WhileStackingException() bool {
	return fs&arm.SCB_CFSR_STKERR != 0
}

// DuringFPLazyStatePres: A bus fault occurred during floating-point lazy state
// preservation
func (fs BusFaultStatus) DuringFPLazyStatePres() bool {
	return fs&arm.SCB_CFSR_LSPERR != 0
}

// UndefinedInstruction: the processor has attempted to execute an undefined
// instruction
//
// "When this bit is set to 1, the PC value stacked for the exception return
// points to the undefined instruction.
//
// An undefined instruction is an instruction that the processor cannot decode."
func (fs UsageFaultStatus) UndefinedInstruction() bool {
	return fs&arm.SCB_CFSR_UNDEFINSTR != 0
}

// IllegalUseOfEPSR: the processor has attempted to execute an instruction that
// makes illegal use of the EPSR
//
// "When this bit is set to 1, the PC value stacked for the exception return
// points to the instruction that attempted the illegal use of the EPSR.
//
// This bit is not set to 1 if an undefined instruction uses the EPSR."
func (fs UsageFaultStatus) IllegalUseOfEPSR() bool {
	return fs&arm.SCB_CFSR_INVSTATE != 0
}

// IllegalExceptionReturn: the processor has attempted an illegal load of
// EXC_RETURN to the PC
//
// "When this bit is set to 1, the PC value stacked for the exception return
// points to the instruction that tried to perform the illegal load of the PC."
func (fs UsageFaultStatus) IllegalExceptionReturn() bool {
	return fs&arm.SCB_CFSR_INVPC != 0
}

// AttemptedToAccessCoprocessor: the processor has attempted to access a
// coprocessor
func (fs UsageFaultStatus) AttemptedToAccessCoprocessor() bool {
	return fs&arm.SCB_CFSR_NOCP != 0
}

// UnalignedMemoryAccess: the processor has made an unaligned memory access
//
// "Enable trapping of unaligned accesses by setting the UNALIGN_TRP bit in the
// CCR to 1.
//
// Unaligned LDM, STM, LDRD, and STRD instructions always fault irrespective of
// the setting of UNALIGN_TRP."
func (fs UsageFaultStatus) UnalignedMemoryAccess() bool {
	return fs&arm.SCB_CFSR_UNALIGNED != 0
}

// DivideByZero: the processor has executed an SDIV or UDIV instruction with a
// divisor of 0
//
// "When the processor sets this bit to 1, the PC value stacked for the
// exception return points to the instruction that performed the divide by zero.
//
// Enable trapping of divide by zero by setting the DIV_0_TRP bit in the CCR to
// 1."
func (fs UsageFaultStatus) DivideByZero() bool {
	return fs&arm.SCB_CFSR_DIVBYZERO != 0
}

// Address returns the MemManage Fault Address Register if the fault status
// indicates the address is valid.
//
// "If a MemManage fault occurs and is escalated to a HardFault because of
// priority, the HardFault handler must set this bit to 0. This prevents
// problems on return to a stacked active MemManage fault handler whose MMAR
// value has been overwritten."
func (fs MemFaultStatus) Address() (uintptr, bool) {
	if fs&arm.SCB_CFSR_MMARVALID == 0 {
		return 0, false
	} else {
		return uintptr(arm.SCB.MMFAR.Get()), true
	}
}

// Address returns the BusFault Address Register if the fault status
// indicates the address is valid.
//
// "The processor sets this bit to 1 after a BusFault where the address is
// known. Other faults can set this bit to 0, such as a MemManage fault
// occurring later.
//
// If a BusFault occurs and is escalated to a hard fault because of priority,
// the hard fault handler must set this bit to 0. This prevents problems if
// returning to a stacked active BusFault handler whose BFAR value has been
// overwritten.""
func (fs BusFaultStatus) Address() (uintptr, bool) {
	if fs&arm.SCB_CFSR_BFARVALID == 0 {
		return 0, false
	} else {
		return uintptr(arm.SCB.BFAR.Get()), true
	}
}
