// +build cortexm3 cortexm4

package arm

const (
	SCB_CFSR_IACCVIOL    = 0x1
	SCB_CFSR_DACCVIOL    = 0x2
	SCB_CFSR_MUNSTKERR   = 0x8
	SCB_CFSR_MSTKERR     = 0x10
	SCB_CFSR_MLSPERR     = 0x20
	SCB_CFSR_MMFARVALID  = 0x80
	SCB_CFSR_IBUSERR     = 0x100
	SCB_CFSR_PRECISERR   = 0x200
	SCB_CFSR_IMPRECISERR = 0x400
	SCB_CFSR_UNSTKERR    = 0x800
	SCB_CFSR_STKERR      = 0x1000
	SCB_CFSR_LSPERR      = 0x2000
	SCB_CFSR_BFARVALID   = 0x8000
	SCB_CFSR_UNDEFINSTR  = 0x10000
	SCB_CFSR_INVSTATE    = 0x20000
	SCB_CFSR_INVPC       = 0x40000
	SCB_CFSR_NOCP        = 0x80000
	SCB_CFSR_UNALIGNED   = 0x1000000
	SCB_CFSR_DIVBYZERO   = 0x2000000

	SCB_CFSR_KnownFault = SCB_CFSR_IACCVIOL | SCB_CFSR_DACCVIOL | SCB_CFSR_MUNSTKERR | SCB_CFSR_MSTKERR | SCB_CFSR_MLSPERR | SCB_CFSR_IBUSERR | SCB_CFSR_PRECISERR | SCB_CFSR_IMPRECISERR | SCB_CFSR_UNSTKERR | SCB_CFSR_STKERR | SCB_CFSR_LSPERR | SCB_CFSR_UNDEFINSTR | SCB_CFSR_INVSTATE | SCB_CFSR_INVPC | SCB_CFSR_NOCP | SCB_CFSR_UNALIGNED | SCB_CFSR_DIVBYZERO
)

// GetFaultStatus reads the System Control Block Configurable Fault Status
// Register and returns it as a FaultStatus.
func GetFaultStatus() FaultStatus { return FaultStatus(SCB.CFSR.Get()) }

// FaultStatus aids in decoding the CFSR.
type FaultStatus uint32

// MemFaultStatus aids in decoding the MFSR.
type MemFaultStatus uint32

// BusFaultStatus aids in decoding the BFSR.
type BusFaultStatus uint32

// UsageFaultStatus aids in decoding the UFSR.
type UsageFaultStatus uint32

// Mem returns `fs` as a MemFaultStatus
func (fs FaultStatus) Mem() MemFaultStatus { return MemFaultStatus(fs) }

// Bus returns `fs` as a BusFaultStatus
func (fs FaultStatus) Bus() BusFaultStatus { return BusFaultStatus(fs) }

// Usage returns `fs` as a UsageFaultStatus
func (fs FaultStatus) Usage() UsageFaultStatus { return UsageFaultStatus(fs) }

// Unknown returns true if the cause of the fault is not known.
func (fs FaultStatus) Unknown() bool { return fs&SCB_CFSR_KnownFault == 0 }

// InstructionAccessViolation indicates a memory access violation error occured
// when fetching an instruciton. This is due to a memory region being
// inaccessible or non-executable for the current priviledge level.
//
// The stacked PC value may point to the faulting instruction.
func (fs MemFaultStatus) InstructionAccessViolation() bool { return fs&SCB_CFSR_IACCVIOL != 0 }

// DataAccessViolation indicates a memory access violation occured when fetching
// or storing data. This is due to attempting to read or write a memory region
// that cannot be read or written, respectively, by the current priviledge
// level.
//
// The MMFAR may be loaded with the faulting address. The stacked PC value may
// point to the faulting instruction.
func (fs MemFaultStatus) DataAccessViolation() bool { return fs&SCB_CFSR_DACCVIOL != 0 }

// WhileUnstackingException indicates a memory fault occured while unstacking
// (after an exception).
//
// The stacked PC is unchanged.
func (fs MemFaultStatus) WhileUnstackingException() bool { return fs&SCB_CFSR_MUNSTKERR != 0 }

// WhileStackingException indicates a memory fault occured while stacking (prior
// to exception handling).
func (fs MemFaultStatus) WhileStackingException() bool { return fs&SCB_CFSR_MSTKERR != 0 }

// WhileLazyStackingException indicates a memory fault occured while lazy
// stacking (deferred stacking of floating point registers).
func (fs MemFaultStatus) WhileLazyStackingException() bool { return fs&SCB_CFSR_MLSPERR != 0 }

// InstructionBusError indicates a bus error occured when fetching an
// instruciton. This is caused by branching to an invalid memory location.
func (fs BusFaultStatus) InstructionBusError() bool { return fs&SCB_CFSR_IBUSERR != 0 }

// PreciseDataBusError indicates a bus error occured when fetching or storing
// data from an unbuffered memory region.
//
// The BFAR is loaded with the faulting address. The stacked PC value may point
// to the faulting instruction.
func (fs BusFaultStatus) PreciseDataBusError() bool { return fs&SCB_CFSR_PRECISERR != 0 }

// ImpreciseDataBusError indicates a bus error occured when fetching or storing
// data from a buffered memory region. Program execution has likely continued
// past the faulting instruction.
//
// The BFAR is not valid. The stacked PC value does not point to the faulting
// instruction.
func (fs BusFaultStatus) ImpreciseDataBusError() bool { return fs&SCB_CFSR_IMPRECISERR != 0 }

// WhileUnstackingException indiciates a bus error occured while unstacking
// (after an exception).
//
// The stacked PC is unchanged.
func (fs BusFaultStatus) WhileUnstackingException() bool { return fs&SCB_CFSR_UNSTKERR != 0 }

// WhileStackingException indicates a bus error occured while stacking (prior to
// exception handling).
func (fs BusFaultStatus) WhileStackingException() bool { return fs&SCB_CFSR_STKERR != 0 }

// WhileLazyStackingException indicates a bus error occured while lazy stacking
// (deferred stacking of floating point registers).
func (fs BusFaultStatus) WhileLazyStackingException() bool { return fs&SCB_CFSR_LSPERR != 0 }

// UndefinedInstruction indicates an attempt to execute an undefined
// instruction.
//
// The stacked PC value may point to the faulting instruction.
func (fs UsageFaultStatus) UndefinedInstruction() bool { return fs&SCB_CFSR_UNDEFINSTR != 0 }

// InvalidState indicates an attempt to switch to an invalid state.
//
// The stacked PC value may point to the faulting instruction.
func (fs UsageFaultStatus) InvalidState() bool { return fs&SCB_CFSR_INVSTATE != 0 }

// IllegalExceptionReturn indicates an attempt to execute an exception return by
// loading the PC with an illegal EXC_RETURN value.
//
// The stacked PC value may point to the faulting instruction.
func (fs UsageFaultStatus) IllegalExceptionReturn() bool { return fs&SCB_CFSR_INVPC != 0 }

// CoprocessorAccessViolation indicates an invalid attempt to execute a
// coprocessor instruction.
func (fs UsageFaultStatus) CoprocessorAccessViolation() bool { return fs&SCB_CFSR_NOCP != 0 }

// UnalignedMemoryAccess indicates an unaligned memory access. SIMD instructions
// must always be aligned. Trapping of all unaligned accesses is enabled by
// setting the UNALIGN_TRP bit in the CCR.
func (fs UsageFaultStatus) UnalignedMemoryAccess() bool { return fs&SCB_CFSR_UNALIGNED != 0 }

// DivideByZero indicates a attempt to execute an SDIV or UDIV instruction with
// a divisor of 0. Trapping of devide-by-zero is enabled by setting the
// DIV_0_TRP bit in the CCR.
func (fs UsageFaultStatus) DivideByZero() bool { return fs&SCB_CFSR_DIVBYZERO != 0 }

// Address returns the MemManage Fault Address Register if the fault status
// indicates the address is valid.
func (fs MemFaultStatus) Address() (uintptr, bool) {
	// check the status after reading the address
	addr := uintptr(SCB.MMFAR.Get())
	ok := fs&SCB_CFSR_MMFARVALID == 0
	return addr, ok
}

// Address returns the BusFault Address Register if the fault status indicates
// the address is valid.
func (fs BusFaultStatus) Address() (uintptr, bool) {
	// check the status after reading the address
	addr := uintptr(SCB.BFAR.Get())
	ok := fs&SCB_CFSR_BFARVALID == 0
	return addr, ok
}
