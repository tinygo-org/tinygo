// Hand created file. DO NOT DELETE.
// Cortex-M System Control Block-related definitions.

//go:build cortexm

package arm

import (
	"runtime/volatile"
	"unsafe"
)

const SCB_BASE = SCS_BASE + 0x0D00

// System Control Block (SCB)
//
// SCB_Type provides the definitions for the System Control Block Registers.
type SCB_Type struct {
	CPUID volatile.Register32 // 0xD00: CPUID Base Register
	ICSR  volatile.Register32 // 0xD04: Interrupt Control and State Register
	VTOR  volatile.Register32 // 0xD08: Vector Table Offset Register
	AIRCR volatile.Register32 // 0xD0C: Application Interrupt and Reset Control Register
	SCR   volatile.Register32 // 0xD10: System Control Register
	CCR   volatile.Register32 // 0xD14: Configuration and Control Register
	SHPR1 volatile.Register32 // 0xD18: System Handler Priority Register 1 (Cortex-M3/M33/M4/M7 only)
	SHPR2 volatile.Register32 // 0xD1C: System Handler Priority Register 2
	SHPR3 volatile.Register32 // 0xD20: System Handler Priority Register 3
	// the following are only applicable for Cortex-M3/M33/M4/M7
	SHCSR volatile.Register32    // 0xD24: System Handler Control and State Register
	CFSR  volatile.Register32    // 0xD28: Configurable Fault Status Register
	HFSR  volatile.Register32    // 0xD2C: HardFault Status Register
	DFSR  volatile.Register32    // 0xD30: Debug Fault Status Register
	MMFAR volatile.Register32    // 0xD34: MemManage Fault Address Register
	BFAR  volatile.Register32    // 0xD38: BusFault Address Register
	AFSR  volatile.Register32    // 0xD3C: Auxiliary Fault Status Register
	PFR   [2]volatile.Register32 // 0xD40: Processor Feature Register
	DFR   volatile.Register32    // 0xD48: Debug Feature Register
	ADR   volatile.Register32    // 0xD4C: Auxiliary Feature Register
	MMFR  [4]volatile.Register32 // 0xD50: Memory Model Feature Register
	ISAR  [5]volatile.Register32 // 0xD60: Instruction Set Attributes Register
	_     [5]uint32              // reserved
	CPACR volatile.Register32    // 0xD88: Coprocessor Access Control Register

}

var SCB = (*SCB_Type)(unsafe.Pointer(uintptr(SCB_BASE)))

// SystemReset performs a hard system reset.
func SystemReset() {
	SCB.AIRCR.Set((0x5FA << SCB_AIRCR_VECTKEY_Pos) | SCB_AIRCR_SYSRESETREQ_Msk)
	for {
		Asm("wfi")
	}
}

const (
	// CPUID: CPUID Base Register
	SCB_CPUID_REVISION_Pos     = 0x0        // Position of REVISION field.
	SCB_CPUID_REVISION_Msk     = 0xf        // Bit mask of REVISION field.
	SCB_CPUID_PARTNO_Pos       = 0x4        // Position of PARTNO field.
	SCB_CPUID_PARTNO_Msk       = 0xfff0     // Bit mask of PARTNO field.
	SCB_CPUID_ARCHITECTURE_Pos = 0x10       // Position of ARCHITECTURE field.
	SCB_CPUID_ARCHITECTURE_Msk = 0xf0000    // Bit mask of ARCHITECTURE field.
	SCB_CPUID_VARIANT_Pos      = 0x14       // Position of VARIANT field.
	SCB_CPUID_VARIANT_Msk      = 0xf00000   // Bit mask of VARIANT field.
	SCB_CPUID_IMPLEMENTER_Pos  = 0x18       // Position of IMPLEMENTER field.
	SCB_CPUID_IMPLEMENTER_Msk  = 0xff000000 // Bit mask of IMPLEMENTER field.

	// ICSR: Interrupt Control and State Register
	SCB_ICSR_VECTACTIVE_Pos          = 0x0        // Position of VECTACTIVE field.
	SCB_ICSR_VECTACTIVE_Msk          = 0x1ff      // Bit mask of VECTACTIVE field.
	SCB_ICSR_RETTOBASE_Pos           = 0xb        // Position of RETTOBASE field.
	SCB_ICSR_RETTOBASE_Msk           = 0x800      // Bit mask of RETTOBASE field.
	SCB_ICSR_RETTOBASE               = 0x800      // Bit RETTOBASE.
	SCB_ICSR_RETTOBASE_RETTOBASE_0   = 0x0        // there are preempted active exceptions to execute
	SCB_ICSR_RETTOBASE_RETTOBASE_1   = 0x1        // there are no active exceptions, or the currently-executing exception is the only active exception
	SCB_ICSR_VECTPENDING_Pos         = 0xc        // Position of VECTPENDING field.
	SCB_ICSR_VECTPENDING_Msk         = 0x1ff000   // Bit mask of VECTPENDING field.
	SCB_ICSR_ISRPENDING_Pos          = 0x16       // Position of ISRPENDING field.
	SCB_ICSR_ISRPENDING_Msk          = 0x400000   // Bit mask of ISRPENDING field.
	SCB_ICSR_ISRPENDING              = 0x400000   // Bit ISRPENDING.
	SCB_ICSR_ISRPENDING_ISRPENDING_0 = 0x0        // No external interrupt pending.
	SCB_ICSR_ISRPENDING_ISRPENDING_1 = 0x1        // External interrupt pending.
	SCB_ICSR_PENDSTCLR_Pos           = 0x19       // Position of PENDSTCLR field.
	SCB_ICSR_PENDSTCLR_Msk           = 0x2000000  // Bit mask of PENDSTCLR field.
	SCB_ICSR_PENDSTCLR               = 0x2000000  // Bit PENDSTCLR.
	SCB_ICSR_PENDSTCLR_PENDSTCLR_0   = 0x0        // no effect
	SCB_ICSR_PENDSTCLR_PENDSTCLR_1   = 0x1        // removes the pending state from the SysTick exception
	SCB_ICSR_PENDSTSET_Pos           = 0x1a       // Position of PENDSTSET field.
	SCB_ICSR_PENDSTSET_Msk           = 0x4000000  // Bit mask of PENDSTSET field.
	SCB_ICSR_PENDSTSET               = 0x4000000  // Bit PENDSTSET.
	SCB_ICSR_PENDSTSET_PENDSTSET_0   = 0x0        // write: no effect; read: SysTick exception is not pending
	SCB_ICSR_PENDSTSET_PENDSTSET_1   = 0x1        // write: changes SysTick exception state to pending; read: SysTick exception is pending
	SCB_ICSR_PENDSVCLR_Pos           = 0x1b       // Position of PENDSVCLR field.
	SCB_ICSR_PENDSVCLR_Msk           = 0x8000000  // Bit mask of PENDSVCLR field.
	SCB_ICSR_PENDSVCLR               = 0x8000000  // Bit PENDSVCLR.
	SCB_ICSR_PENDSVCLR_PENDSVCLR_0   = 0x0        // no effect
	SCB_ICSR_PENDSVCLR_PENDSVCLR_1   = 0x1        // removes the pending state from the PendSV exception
	SCB_ICSR_PENDSVSET_Pos           = 0x1c       // Position of PENDSVSET field.
	SCB_ICSR_PENDSVSET_Msk           = 0x10000000 // Bit mask of PENDSVSET field.
	SCB_ICSR_PENDSVSET               = 0x10000000 // Bit PENDSVSET.
	SCB_ICSR_PENDSVSET_PENDSVSET_0   = 0x0        // write: no effect; read: PendSV exception is not pending
	SCB_ICSR_PENDSVSET_PENDSVSET_1   = 0x1        // write: changes PendSV exception state to pending; read: PendSV exception is pending
	SCB_ICSR_NMIPENDSET_Pos          = 0x1f       // Position of NMIPENDSET field.
	SCB_ICSR_NMIPENDSET_Msk          = 0x80000000 // Bit mask of NMIPENDSET field.
	SCB_ICSR_NMIPENDSET              = 0x80000000 // Bit NMIPENDSET.
	SCB_ICSR_NMIPENDSET_NMIPENDSET_0 = 0x0        // write: no effect; read: NMI exception is not pending
	SCB_ICSR_NMIPENDSET_NMIPENDSET_1 = 0x1        // write: changes NMI exception state to pending; read: NMI exception is pending

	// VTOR: Vector Table Offset Register
	SCB_VTOR_TBLOFF_Pos = 0x7        // Position of TBLOFF field.
	SCB_VTOR_TBLOFF_Msk = 0xffffff80 // Bit mask of TBLOFF field.

	// AIRCR: Application Interrupt and Reset Control Register
	SCB_AIRCR_VECTRESET_Pos                 = 0x0        // Position of VECTRESET field.
	SCB_AIRCR_VECTRESET_Msk                 = 0x1        // Bit mask of VECTRESET field.
	SCB_AIRCR_VECTRESET                     = 0x1        // Bit VECTRESET.
	SCB_AIRCR_VECTRESET_VECTRESET_0         = 0x0        // No change
	SCB_AIRCR_VECTRESET_VECTRESET_1         = 0x1        // Causes a local system reset
	SCB_AIRCR_VECTCLRACTIVE_Pos             = 0x1        // Position of VECTCLRACTIVE field.
	SCB_AIRCR_VECTCLRACTIVE_Msk             = 0x2        // Bit mask of VECTCLRACTIVE field.
	SCB_AIRCR_VECTCLRACTIVE                 = 0x2        // Bit VECTCLRACTIVE.
	SCB_AIRCR_VECTCLRACTIVE_VECTCLRACTIVE_0 = 0x0        // No change
	SCB_AIRCR_VECTCLRACTIVE_VECTCLRACTIVE_1 = 0x1        // Clears all active state information for fixed and configurable exceptions
	SCB_AIRCR_SYSRESETREQ_Pos               = 0x2        // Position of SYSRESETREQ field.
	SCB_AIRCR_SYSRESETREQ_Msk               = 0x4        // Bit mask of SYSRESETREQ field.
	SCB_AIRCR_SYSRESETREQ                   = 0x4        // Bit SYSRESETREQ.
	SCB_AIRCR_SYSRESETREQ_SYSRESETREQ_0     = 0x0        // no system reset request
	SCB_AIRCR_SYSRESETREQ_SYSRESETREQ_1     = 0x1        // asserts a signal to the outer system that requests a reset
	SCB_AIRCR_PRIGROUP_Pos                  = 0x8        // Position of PRIGROUP field.
	SCB_AIRCR_PRIGROUP_Msk                  = 0x700      // Bit mask of PRIGROUP field.
	SCB_AIRCR_ENDIANNESS_Pos                = 0xf        // Position of ENDIANNESS field.
	SCB_AIRCR_ENDIANNESS_Msk                = 0x8000     // Bit mask of ENDIANNESS field.
	SCB_AIRCR_ENDIANNESS                    = 0x8000     // Bit ENDIANNESS.
	SCB_AIRCR_ENDIANNESS_ENDIANNESS_0       = 0x0        // Little-endian
	SCB_AIRCR_ENDIANNESS_ENDIANNESS_1       = 0x1        // Big-endian
	SCB_AIRCR_VECTKEY_Pos                   = 0x10       // Position of VECTKEY field.
	SCB_AIRCR_VECTKEY_Msk                   = 0xffff0000 // Bit mask of VECTKEY field.

	// SCR: System Control Register
	SCB_SCR_SLEEPONEXIT_Pos           = 0x1  // Position of SLEEPONEXIT field.
	SCB_SCR_SLEEPONEXIT_Msk           = 0x2  // Bit mask of SLEEPONEXIT field.
	SCB_SCR_SLEEPONEXIT               = 0x2  // Bit SLEEPONEXIT.
	SCB_SCR_SLEEPONEXIT_SLEEPONEXIT_0 = 0x0  // o not sleep when returning to Thread mode
	SCB_SCR_SLEEPONEXIT_SLEEPONEXIT_1 = 0x1  // enter sleep, or deep sleep, on return from an ISR
	SCB_SCR_SLEEPDEEP_Pos             = 0x2  // Position of SLEEPDEEP field.
	SCB_SCR_SLEEPDEEP_Msk             = 0x4  // Bit mask of SLEEPDEEP field.
	SCB_SCR_SLEEPDEEP                 = 0x4  // Bit SLEEPDEEP.
	SCB_SCR_SLEEPDEEP_SLEEPDEEP_0     = 0x0  // sleep
	SCB_SCR_SLEEPDEEP_SLEEPDEEP_1     = 0x1  // deep sleep
	SCB_SCR_SEVONPEND_Pos             = 0x4  // Position of SEVONPEND field.
	SCB_SCR_SEVONPEND_Msk             = 0x10 // Bit mask of SEVONPEND field.
	SCB_SCR_SEVONPEND                 = 0x10 // Bit SEVONPEND.
	SCB_SCR_SEVONPEND_SEVONPEND_0     = 0x0  // only enabled interrupts or events can wakeup the processor, disabled interrupts are excluded
	SCB_SCR_SEVONPEND_SEVONPEND_1     = 0x1  // enabled events and all interrupts, including disabled interrupts, can wakeup the processor

	// CCR: Configuration and Control Register
	SCB_CCR_NONBASETHRDENA_Pos              = 0x0     // Position of NONBASETHRDENA field.
	SCB_CCR_NONBASETHRDENA_Msk              = 0x1     // Bit mask of NONBASETHRDENA field.
	SCB_CCR_NONBASETHRDENA                  = 0x1     // Bit NONBASETHRDENA.
	SCB_CCR_NONBASETHRDENA_NONBASETHRDENA_0 = 0x0     // processor can enter Thread mode only when no exception is active
	SCB_CCR_NONBASETHRDENA_NONBASETHRDENA_1 = 0x1     // processor can enter Thread mode from any level under the control of an EXC_RETURN value
	SCB_CCR_USERSETMPEND_Pos                = 0x1     // Position of USERSETMPEND field.
	SCB_CCR_USERSETMPEND_Msk                = 0x2     // Bit mask of USERSETMPEND field.
	SCB_CCR_USERSETMPEND                    = 0x2     // Bit USERSETMPEND.
	SCB_CCR_USERSETMPEND_USERSETMPEND_0     = 0x0     // disable
	SCB_CCR_USERSETMPEND_USERSETMPEND_1     = 0x1     // enable
	SCB_CCR_UNALIGN_TRP_Pos                 = 0x3     // Position of UNALIGN_TRP field.
	SCB_CCR_UNALIGN_TRP_Msk                 = 0x8     // Bit mask of UNALIGN_TRP field.
	SCB_CCR_UNALIGN_TRP                     = 0x8     // Bit UNALIGN_TRP.
	SCB_CCR_UNALIGN_TRP_UNALIGN_TRP_0       = 0x0     // do not trap unaligned halfword and word accesses
	SCB_CCR_UNALIGN_TRP_UNALIGN_TRP_1       = 0x1     // trap unaligned halfword and word accesses
	SCB_CCR_DIV_0_TRP_Pos                   = 0x4     // Position of DIV_0_TRP field.
	SCB_CCR_DIV_0_TRP_Msk                   = 0x10    // Bit mask of DIV_0_TRP field.
	SCB_CCR_DIV_0_TRP                       = 0x10    // Bit DIV_0_TRP.
	SCB_CCR_DIV_0_TRP_DIV_0_TRP_0           = 0x0     // do not trap divide by 0
	SCB_CCR_DIV_0_TRP_DIV_0_TRP_1           = 0x1     // trap divide by 0
	SCB_CCR_BFHFNMIGN_Pos                   = 0x8     // Position of BFHFNMIGN field.
	SCB_CCR_BFHFNMIGN_Msk                   = 0x100   // Bit mask of BFHFNMIGN field.
	SCB_CCR_BFHFNMIGN                       = 0x100   // Bit BFHFNMIGN.
	SCB_CCR_BFHFNMIGN_BFHFNMIGN_0           = 0x0     // data bus faults caused by load and store instructions cause a lock-up
	SCB_CCR_BFHFNMIGN_BFHFNMIGN_1           = 0x1     // handlers running at priority -1 and -2 ignore data bus faults caused by load and store instructions
	SCB_CCR_STKALIGN_Pos                    = 0x9     // Position of STKALIGN field.
	SCB_CCR_STKALIGN_Msk                    = 0x200   // Bit mask of STKALIGN field.
	SCB_CCR_STKALIGN                        = 0x200   // Bit STKALIGN.
	SCB_CCR_STKALIGN_STKALIGN_0             = 0x0     // 4-byte aligned
	SCB_CCR_STKALIGN_STKALIGN_1             = 0x1     // 8-byte aligned
	SCB_CCR_DC_Pos                          = 0x10    // Position of DC field.
	SCB_CCR_DC_Msk                          = 0x10000 // Bit mask of DC field.
	SCB_CCR_DC                              = 0x10000 // Bit DC.
	SCB_CCR_DC_DC_0                         = 0x0     // L1 data cache disabled
	SCB_CCR_DC_DC_1                         = 0x1     // L1 data cache enabled
	SCB_CCR_IC_Pos                          = 0x11    // Position of IC field.
	SCB_CCR_IC_Msk                          = 0x20000 // Bit mask of IC field.
	SCB_CCR_IC                              = 0x20000 // Bit IC.
	SCB_CCR_IC_IC_0                         = 0x0     // L1 instruction cache disabled
	SCB_CCR_IC_IC_1                         = 0x1     // L1 instruction cache enabled
	SCB_CCR_BP_Pos                          = 0x12    // Position of BP field.
	SCB_CCR_BP_Msk                          = 0x40000 // Bit mask of BP field.
	SCB_CCR_BP                              = 0x40000 // Bit BP.

	// SHPR1: System Handler Priority Register 1
	SCB_SHPR1_PRI_4_Pos = 0x0      // Position of PRI_4 field.
	SCB_SHPR1_PRI_4_Msk = 0xff     // Bit mask of PRI_4 field.
	SCB_SHPR1_PRI_5_Pos = 0x8      // Position of PRI_5 field.
	SCB_SHPR1_PRI_5_Msk = 0xff00   // Bit mask of PRI_5 field.
	SCB_SHPR1_PRI_6_Pos = 0x10     // Position of PRI_6 field.
	SCB_SHPR1_PRI_6_Msk = 0xff0000 // Bit mask of PRI_6 field.

	// SHPR2: System Handler Priority Register 2
	SCB_SHPR2_PRI_11_Pos = 0x18       // Position of PRI_11 field.
	SCB_SHPR2_PRI_11_Msk = 0xff000000 // Bit mask of PRI_11 field.

	// SHPR3: System Handler Priority Register 3
	SCB_SHPR3_PRI_14_Pos = 0x10       // Position of PRI_14 field.
	SCB_SHPR3_PRI_14_Msk = 0xff0000   // Bit mask of PRI_14 field.
	SCB_SHPR3_PRI_15_Pos = 0x18       // Position of PRI_15 field.
	SCB_SHPR3_PRI_15_Msk = 0xff000000 // Bit mask of PRI_15 field.

	// SHCSR: System Handler Control and State Register
	SCB_SHCSR_MEMFAULTACT_Pos                 = 0x0     // Position of MEMFAULTACT field.
	SCB_SHCSR_MEMFAULTACT_Msk                 = 0x1     // Bit mask of MEMFAULTACT field.
	SCB_SHCSR_MEMFAULTACT                     = 0x1     // Bit MEMFAULTACT.
	SCB_SHCSR_MEMFAULTACT_MEMFAULTACT_0       = 0x0     // exception is not active
	SCB_SHCSR_MEMFAULTACT_MEMFAULTACT_1       = 0x1     // exception is active
	SCB_SHCSR_BUSFAULTACT_Pos                 = 0x1     // Position of BUSFAULTACT field.
	SCB_SHCSR_BUSFAULTACT_Msk                 = 0x2     // Bit mask of BUSFAULTACT field.
	SCB_SHCSR_BUSFAULTACT                     = 0x2     // Bit BUSFAULTACT.
	SCB_SHCSR_BUSFAULTACT_BUSFAULTACT_0       = 0x0     // exception is not active
	SCB_SHCSR_BUSFAULTACT_BUSFAULTACT_1       = 0x1     // exception is active
	SCB_SHCSR_USGFAULTACT_Pos                 = 0x3     // Position of USGFAULTACT field.
	SCB_SHCSR_USGFAULTACT_Msk                 = 0x8     // Bit mask of USGFAULTACT field.
	SCB_SHCSR_USGFAULTACT                     = 0x8     // Bit USGFAULTACT.
	SCB_SHCSR_USGFAULTACT_USGFAULTACT_0       = 0x0     // exception is not active
	SCB_SHCSR_USGFAULTACT_USGFAULTACT_1       = 0x1     // exception is active
	SCB_SHCSR_SVCALLACT_Pos                   = 0x7     // Position of SVCALLACT field.
	SCB_SHCSR_SVCALLACT_Msk                   = 0x80    // Bit mask of SVCALLACT field.
	SCB_SHCSR_SVCALLACT                       = 0x80    // Bit SVCALLACT.
	SCB_SHCSR_SVCALLACT_SVCALLACT_0           = 0x0     // exception is not active
	SCB_SHCSR_SVCALLACT_SVCALLACT_1           = 0x1     // exception is active
	SCB_SHCSR_MONITORACT_Pos                  = 0x8     // Position of MONITORACT field.
	SCB_SHCSR_MONITORACT_Msk                  = 0x100   // Bit mask of MONITORACT field.
	SCB_SHCSR_MONITORACT                      = 0x100   // Bit MONITORACT.
	SCB_SHCSR_MONITORACT_MONITORACT_0         = 0x0     // exception is not active
	SCB_SHCSR_MONITORACT_MONITORACT_1         = 0x1     // exception is active
	SCB_SHCSR_PENDSVACT_Pos                   = 0xa     // Position of PENDSVACT field.
	SCB_SHCSR_PENDSVACT_Msk                   = 0x400   // Bit mask of PENDSVACT field.
	SCB_SHCSR_PENDSVACT                       = 0x400   // Bit PENDSVACT.
	SCB_SHCSR_PENDSVACT_PENDSVACT_0           = 0x0     // exception is not active
	SCB_SHCSR_PENDSVACT_PENDSVACT_1           = 0x1     // exception is active
	SCB_SHCSR_SYSTICKACT_Pos                  = 0xb     // Position of SYSTICKACT field.
	SCB_SHCSR_SYSTICKACT_Msk                  = 0x800   // Bit mask of SYSTICKACT field.
	SCB_SHCSR_SYSTICKACT                      = 0x800   // Bit SYSTICKACT.
	SCB_SHCSR_SYSTICKACT_SYSTICKACT_0         = 0x0     // exception is not active
	SCB_SHCSR_SYSTICKACT_SYSTICKACT_1         = 0x1     // exception is active
	SCB_SHCSR_USGFAULTPENDED_Pos              = 0xc     // Position of USGFAULTPENDED field.
	SCB_SHCSR_USGFAULTPENDED_Msk              = 0x1000  // Bit mask of USGFAULTPENDED field.
	SCB_SHCSR_USGFAULTPENDED                  = 0x1000  // Bit USGFAULTPENDED.
	SCB_SHCSR_USGFAULTPENDED_USGFAULTPENDED_0 = 0x0     // exception is not pending
	SCB_SHCSR_USGFAULTPENDED_USGFAULTPENDED_1 = 0x1     // exception is pending
	SCB_SHCSR_MEMFAULTPENDED_Pos              = 0xd     // Position of MEMFAULTPENDED field.
	SCB_SHCSR_MEMFAULTPENDED_Msk              = 0x2000  // Bit mask of MEMFAULTPENDED field.
	SCB_SHCSR_MEMFAULTPENDED                  = 0x2000  // Bit MEMFAULTPENDED.
	SCB_SHCSR_MEMFAULTPENDED_MEMFAULTPENDED_0 = 0x0     // exception is not pending
	SCB_SHCSR_MEMFAULTPENDED_MEMFAULTPENDED_1 = 0x1     // exception is pending
	SCB_SHCSR_BUSFAULTPENDED_Pos              = 0xe     // Position of BUSFAULTPENDED field.
	SCB_SHCSR_BUSFAULTPENDED_Msk              = 0x4000  // Bit mask of BUSFAULTPENDED field.
	SCB_SHCSR_BUSFAULTPENDED                  = 0x4000  // Bit BUSFAULTPENDED.
	SCB_SHCSR_BUSFAULTPENDED_BUSFAULTPENDED_0 = 0x0     // exception is not pending
	SCB_SHCSR_BUSFAULTPENDED_BUSFAULTPENDED_1 = 0x1     // exception is pending
	SCB_SHCSR_SVCALLPENDED_Pos                = 0xf     // Position of SVCALLPENDED field.
	SCB_SHCSR_SVCALLPENDED_Msk                = 0x8000  // Bit mask of SVCALLPENDED field.
	SCB_SHCSR_SVCALLPENDED                    = 0x8000  // Bit SVCALLPENDED.
	SCB_SHCSR_SVCALLPENDED_SVCALLPENDED_0     = 0x0     // exception is not pending
	SCB_SHCSR_SVCALLPENDED_SVCALLPENDED_1     = 0x1     // exception is pending
	SCB_SHCSR_MEMFAULTENA_Pos                 = 0x10    // Position of MEMFAULTENA field.
	SCB_SHCSR_MEMFAULTENA_Msk                 = 0x10000 // Bit mask of MEMFAULTENA field.
	SCB_SHCSR_MEMFAULTENA                     = 0x10000 // Bit MEMFAULTENA.
	SCB_SHCSR_MEMFAULTENA_MEMFAULTENA_0       = 0x0     // disable the exception
	SCB_SHCSR_MEMFAULTENA_MEMFAULTENA_1       = 0x1     // enable the exception
	SCB_SHCSR_BUSFAULTENA_Pos                 = 0x11    // Position of BUSFAULTENA field.
	SCB_SHCSR_BUSFAULTENA_Msk                 = 0x20000 // Bit mask of BUSFAULTENA field.
	SCB_SHCSR_BUSFAULTENA                     = 0x20000 // Bit BUSFAULTENA.
	SCB_SHCSR_BUSFAULTENA_BUSFAULTENA_0       = 0x0     // disable the exception
	SCB_SHCSR_BUSFAULTENA_BUSFAULTENA_1       = 0x1     // enable the exception
	SCB_SHCSR_USGFAULTENA_Pos                 = 0x12    // Position of USGFAULTENA field.
	SCB_SHCSR_USGFAULTENA_Msk                 = 0x40000 // Bit mask of USGFAULTENA field.
	SCB_SHCSR_USGFAULTENA                     = 0x40000 // Bit USGFAULTENA.
	SCB_SHCSR_USGFAULTENA_USGFAULTENA_0       = 0x0     // disable the exception
	SCB_SHCSR_USGFAULTENA_USGFAULTENA_1       = 0x1     // enable the exception

	// CFSR: Configurable Fault Status Register
	SCB_CFSR_IACCVIOL_Pos              = 0x0       // Position of IACCVIOL field.
	SCB_CFSR_IACCVIOL_Msk              = 0x1       // Bit mask of IACCVIOL field.
	SCB_CFSR_IACCVIOL                  = 0x1       // Bit IACCVIOL.
	SCB_CFSR_IACCVIOL_IACCVIOL_0       = 0x0       // no instruction access violation fault
	SCB_CFSR_IACCVIOL_IACCVIOL_1       = 0x1       // the processor attempted an instruction fetch from a location that does not permit execution
	SCB_CFSR_DACCVIOL_Pos              = 0x1       // Position of DACCVIOL field.
	SCB_CFSR_DACCVIOL_Msk              = 0x2       // Bit mask of DACCVIOL field.
	SCB_CFSR_DACCVIOL                  = 0x2       // Bit DACCVIOL.
	SCB_CFSR_DACCVIOL_DACCVIOL_0       = 0x0       // no data access violation fault
	SCB_CFSR_DACCVIOL_DACCVIOL_1       = 0x1       // the processor attempted a load or store at a location that does not permit the operation
	SCB_CFSR_MUNSTKERR_Pos             = 0x3       // Position of MUNSTKERR field.
	SCB_CFSR_MUNSTKERR_Msk             = 0x8       // Bit mask of MUNSTKERR field.
	SCB_CFSR_MUNSTKERR                 = 0x8       // Bit MUNSTKERR.
	SCB_CFSR_MUNSTKERR_MUNSTKERR_0     = 0x0       // no unstacking fault
	SCB_CFSR_MUNSTKERR_MUNSTKERR_1     = 0x1       // unstack for an exception return has caused one or more access violations
	SCB_CFSR_MSTKERR_Pos               = 0x4       // Position of MSTKERR field.
	SCB_CFSR_MSTKERR_Msk               = 0x10      // Bit mask of MSTKERR field.
	SCB_CFSR_MSTKERR                   = 0x10      // Bit MSTKERR.
	SCB_CFSR_MSTKERR_MSTKERR_0         = 0x0       // no stacking fault
	SCB_CFSR_MSTKERR_MSTKERR_1         = 0x1       // stacking for an exception entry has caused one or more access violations
	SCB_CFSR_MLSPERR_Pos               = 0x5       // Position of MLSPERR field.
	SCB_CFSR_MLSPERR_Msk               = 0x20      // Bit mask of MLSPERR field.
	SCB_CFSR_MLSPERR                   = 0x20      // Bit MLSPERR.
	SCB_CFSR_MLSPERR_MLSPERR_0         = 0x0       // No MemManage fault occurred during floating-point lazy state preservation
	SCB_CFSR_MLSPERR_MLSPERR_1         = 0x1       // A MemManage fault occurred during floating-point lazy state preservation
	SCB_CFSR_MMARVALID_Pos             = 0x7       // Position of MMARVALID field.
	SCB_CFSR_MMARVALID_Msk             = 0x80      // Bit mask of MMARVALID field.
	SCB_CFSR_MMARVALID                 = 0x80      // Bit MMARVALID.
	SCB_CFSR_MMARVALID_MMARVALID_0     = 0x0       // value in MMAR is not a valid fault address
	SCB_CFSR_MMARVALID_MMARVALID_1     = 0x1       // MMAR holds a valid fault address
	SCB_CFSR_IBUSERR_Pos               = 0x8       // Position of IBUSERR field.
	SCB_CFSR_IBUSERR_Msk               = 0x100     // Bit mask of IBUSERR field.
	SCB_CFSR_IBUSERR                   = 0x100     // Bit IBUSERR.
	SCB_CFSR_IBUSERR_IBUSERR_0         = 0x0       // no instruction bus error
	SCB_CFSR_IBUSERR_IBUSERR_1         = 0x1       // instruction bus error
	SCB_CFSR_PRECISERR_Pos             = 0x9       // Position of PRECISERR field.
	SCB_CFSR_PRECISERR_Msk             = 0x200     // Bit mask of PRECISERR field.
	SCB_CFSR_PRECISERR                 = 0x200     // Bit PRECISERR.
	SCB_CFSR_PRECISERR_PRECISERR_0     = 0x0       // no precise data bus error
	SCB_CFSR_PRECISERR_PRECISERR_1     = 0x1       // a data bus error has occurred, and the PC value stacked for the exception return points to the instruction that caused the fault
	SCB_CFSR_IMPRECISERR_Pos           = 0xa       // Position of IMPRECISERR field.
	SCB_CFSR_IMPRECISERR_Msk           = 0x400     // Bit mask of IMPRECISERR field.
	SCB_CFSR_IMPRECISERR               = 0x400     // Bit IMPRECISERR.
	SCB_CFSR_IMPRECISERR_IMPRECISERR_0 = 0x0       // no imprecise data bus error
	SCB_CFSR_IMPRECISERR_IMPRECISERR_1 = 0x1       // a data bus error has occurred, but the return address in the stack frame is not related to the instruction that caused the error
	SCB_CFSR_UNSTKERR_Pos              = 0xb       // Position of UNSTKERR field.
	SCB_CFSR_UNSTKERR_Msk              = 0x800     // Bit mask of UNSTKERR field.
	SCB_CFSR_UNSTKERR                  = 0x800     // Bit UNSTKERR.
	SCB_CFSR_UNSTKERR_UNSTKERR_0       = 0x0       // no unstacking fault
	SCB_CFSR_UNSTKERR_UNSTKERR_1       = 0x1       // unstack for an exception return has caused one or more BusFaults
	SCB_CFSR_STKERR_Pos                = 0xc       // Position of STKERR field.
	SCB_CFSR_STKERR_Msk                = 0x1000    // Bit mask of STKERR field.
	SCB_CFSR_STKERR                    = 0x1000    // Bit STKERR.
	SCB_CFSR_STKERR_STKERR_0           = 0x0       // no stacking fault
	SCB_CFSR_STKERR_STKERR_1           = 0x1       // stacking for an exception entry has caused one or more BusFaults
	SCB_CFSR_LSPERR_Pos                = 0xd       // Position of LSPERR field.
	SCB_CFSR_LSPERR_Msk                = 0x2000    // Bit mask of LSPERR field.
	SCB_CFSR_LSPERR                    = 0x2000    // Bit LSPERR.
	SCB_CFSR_LSPERR_LSPERR_0           = 0x0       // No bus fault occurred during floating-point lazy state preservation
	SCB_CFSR_LSPERR_LSPERR_1           = 0x1       // A bus fault occurred during floating-point lazy state preservation
	SCB_CFSR_BFARVALID_Pos             = 0xf       // Position of BFARVALID field.
	SCB_CFSR_BFARVALID_Msk             = 0x8000    // Bit mask of BFARVALID field.
	SCB_CFSR_BFARVALID                 = 0x8000    // Bit BFARVALID.
	SCB_CFSR_BFARVALID_BFARVALID_0     = 0x0       // value in BFAR is not a valid fault address
	SCB_CFSR_BFARVALID_BFARVALID_1     = 0x1       // BFAR holds a valid fault address
	SCB_CFSR_UNDEFINSTR_Pos            = 0x10      // Position of UNDEFINSTR field.
	SCB_CFSR_UNDEFINSTR_Msk            = 0x10000   // Bit mask of UNDEFINSTR field.
	SCB_CFSR_UNDEFINSTR                = 0x10000   // Bit UNDEFINSTR.
	SCB_CFSR_UNDEFINSTR_UNDEFINSTR_0   = 0x0       // no undefined instruction UsageFault
	SCB_CFSR_UNDEFINSTR_UNDEFINSTR_1   = 0x1       // the processor has attempted to execute an undefined instruction
	SCB_CFSR_INVSTATE_Pos              = 0x11      // Position of INVSTATE field.
	SCB_CFSR_INVSTATE_Msk              = 0x20000   // Bit mask of INVSTATE field.
	SCB_CFSR_INVSTATE                  = 0x20000   // Bit INVSTATE.
	SCB_CFSR_INVSTATE_INVSTATE_0       = 0x0       // no invalid state UsageFault
	SCB_CFSR_INVSTATE_INVSTATE_1       = 0x1       // the processor has attempted to execute an instruction that makes illegal use of the EPSR
	SCB_CFSR_INVPC_Pos                 = 0x12      // Position of INVPC field.
	SCB_CFSR_INVPC_Msk                 = 0x40000   // Bit mask of INVPC field.
	SCB_CFSR_INVPC                     = 0x40000   // Bit INVPC.
	SCB_CFSR_INVPC_INVPC_0             = 0x0       // no invalid PC load UsageFault
	SCB_CFSR_INVPC_INVPC_1             = 0x1       // the processor has attempted an illegal load of EXC_RETURN to the PC
	SCB_CFSR_NOCP_Pos                  = 0x13      // Position of NOCP field.
	SCB_CFSR_NOCP_Msk                  = 0x80000   // Bit mask of NOCP field.
	SCB_CFSR_NOCP                      = 0x80000   // Bit NOCP.
	SCB_CFSR_NOCP_NOCP_0               = 0x0       // no UsageFault caused by attempting to access a coprocessor
	SCB_CFSR_NOCP_NOCP_1               = 0x1       // the processor has attempted to access a coprocessor
	SCB_CFSR_UNALIGNED_Pos             = 0x18      // Position of UNALIGNED field.
	SCB_CFSR_UNALIGNED_Msk             = 0x1000000 // Bit mask of UNALIGNED field.
	SCB_CFSR_UNALIGNED                 = 0x1000000 // Bit UNALIGNED.
	SCB_CFSR_UNALIGNED_UNALIGNED_0     = 0x0       // no unaligned access fault, or unaligned access trapping not enabled
	SCB_CFSR_UNALIGNED_UNALIGNED_1     = 0x1       // the processor has made an unaligned memory access
	SCB_CFSR_DIVBYZERO_Pos             = 0x19      // Position of DIVBYZERO field.
	SCB_CFSR_DIVBYZERO_Msk             = 0x2000000 // Bit mask of DIVBYZERO field.
	SCB_CFSR_DIVBYZERO                 = 0x2000000 // Bit DIVBYZERO.
	SCB_CFSR_DIVBYZERO_DIVBYZERO_0     = 0x0       // no divide by zero fault, or divide by zero trapping not enabled
	SCB_CFSR_DIVBYZERO_DIVBYZERO_1     = 0x1       // the processor has executed an SDIV or UDIV instruction with a divisor of 0

	// HFSR: HardFault Status register
	SCB_HFSR_VECTTBL_Pos         = 0x1        // Position of VECTTBL field.
	SCB_HFSR_VECTTBL_Msk         = 0x2        // Bit mask of VECTTBL field.
	SCB_HFSR_VECTTBL             = 0x2        // Bit VECTTBL.
	SCB_HFSR_VECTTBL_VECTTBL_0   = 0x0        // no BusFault on vector table read
	SCB_HFSR_VECTTBL_VECTTBL_1   = 0x1        // BusFault on vector table read
	SCB_HFSR_FORCED_Pos          = 0x1e       // Position of FORCED field.
	SCB_HFSR_FORCED_Msk          = 0x40000000 // Bit mask of FORCED field.
	SCB_HFSR_FORCED              = 0x40000000 // Bit FORCED.
	SCB_HFSR_FORCED_FORCED_0     = 0x0        // no forced HardFault
	SCB_HFSR_FORCED_FORCED_1     = 0x1        // forced HardFault
	SCB_HFSR_DEBUGEVT_Pos        = 0x1f       // Position of DEBUGEVT field.
	SCB_HFSR_DEBUGEVT_Msk        = 0x80000000 // Bit mask of DEBUGEVT field.
	SCB_HFSR_DEBUGEVT            = 0x80000000 // Bit DEBUGEVT.
	SCB_HFSR_DEBUGEVT_DEBUGEVT_0 = 0x0        // No Debug event has occurred.
	SCB_HFSR_DEBUGEVT_DEBUGEVT_1 = 0x1        // Debug event has occurred. The Debug Fault Status Register has been updated.

	// DFSR: Debug Fault Status Register
	SCB_DFSR_HALTED_Pos          = 0x0  // Position of HALTED field.
	SCB_DFSR_HALTED_Msk          = 0x1  // Bit mask of HALTED field.
	SCB_DFSR_HALTED              = 0x1  // Bit HALTED.
	SCB_DFSR_HALTED_HALTED_0     = 0x0  // No active halt request debug event
	SCB_DFSR_HALTED_HALTED_1     = 0x1  // Halt request debug event active
	SCB_DFSR_BKPT_Pos            = 0x1  // Position of BKPT field.
	SCB_DFSR_BKPT_Msk            = 0x2  // Bit mask of BKPT field.
	SCB_DFSR_BKPT                = 0x2  // Bit BKPT.
	SCB_DFSR_BKPT_BKPT_0         = 0x0  // No current breakpoint debug event
	SCB_DFSR_BKPT_BKPT_1         = 0x1  // At least one current breakpoint debug event
	SCB_DFSR_DWTTRAP_Pos         = 0x2  // Position of DWTTRAP field.
	SCB_DFSR_DWTTRAP_Msk         = 0x4  // Bit mask of DWTTRAP field.
	SCB_DFSR_DWTTRAP             = 0x4  // Bit DWTTRAP.
	SCB_DFSR_DWTTRAP_DWTTRAP_0   = 0x0  // No current debug events generated by the DWT
	SCB_DFSR_DWTTRAP_DWTTRAP_1   = 0x1  // At least one current debug event generated by the DWT
	SCB_DFSR_VCATCH_Pos          = 0x3  // Position of VCATCH field.
	SCB_DFSR_VCATCH_Msk          = 0x8  // Bit mask of VCATCH field.
	SCB_DFSR_VCATCH              = 0x8  // Bit VCATCH.
	SCB_DFSR_VCATCH_VCATCH_0     = 0x0  // No Vector catch triggered
	SCB_DFSR_VCATCH_VCATCH_1     = 0x1  // Vector catch triggered
	SCB_DFSR_EXTERNAL_Pos        = 0x4  // Position of EXTERNAL field.
	SCB_DFSR_EXTERNAL_Msk        = 0x10 // Bit mask of EXTERNAL field.
	SCB_DFSR_EXTERNAL            = 0x10 // Bit EXTERNAL.
	SCB_DFSR_EXTERNAL_EXTERNAL_0 = 0x0  // No external debug request debug event
	SCB_DFSR_EXTERNAL_EXTERNAL_1 = 0x1  // External debug request debug event

	// MMFAR: MemManage Fault Address Register
	SCB_MMFAR_ADDRESS_Pos = 0x0        // Position of ADDRESS field.
	SCB_MMFAR_ADDRESS_Msk = 0xffffffff // Bit mask of ADDRESS field.

	// BFAR: BusFault Address Register
	SCB_BFAR_ADDRESS_Pos = 0x0        // Position of ADDRESS field.
	SCB_BFAR_ADDRESS_Msk = 0xffffffff // Bit mask of ADDRESS field.
)
