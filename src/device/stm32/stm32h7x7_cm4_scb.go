// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the SCB peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to the Cortex-M4 core only.

// +build stm32h7x7_cm4

package stm32

// We need this type in the "stm32" package namespace for the runtime code that
// is shared by both cores.

import (
	"device/arm"
)

// System Control Block (SCB)
//
// SCB_Type provides the definitions for the System Control Block Registers.
type SCB_Type arm.SCB_Type

var SCB = arm.SCB

var SCB_SHCSR = SCB.SHCSR

const (
	// SCR: System control register
	SCB_SCR_SLEEPONEXIT_Pos = 0x1  // Position of SLEEPONEXIT field.
	SCB_SCR_SLEEPONEXIT_Msk = 0x2  // Bit mask of SLEEPONEXIT field.
	SCB_SCR_SLEEPONEXIT     = 0x2  // Bit SLEEPONEXIT.
	SCB_SCR_SLEEPDEEP_Pos   = 0x2  // Position of SLEEPDEEP field.
	SCB_SCR_SLEEPDEEP_Msk   = 0x4  // Bit mask of SLEEPDEEP field.
	SCB_SCR_SLEEPDEEP       = 0x4  // Bit SLEEPDEEP.
	SCB_SCR_SEVEONPEND_Pos  = 0x4  // Position of SEVEONPEND field.
	SCB_SCR_SEVEONPEND_Msk  = 0x10 // Bit mask of SEVEONPEND field.
	SCB_SCR_SEVEONPEND      = 0x10 // Bit SEVEONPEND.

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

	SCB_SHPR3_PRI_15_Pos = arm.SCB_SHPR3_PRI_15_Pos
	SCB_SHPR3_PRI_14_Pos = arm.SCB_SHPR3_PRI_14_Pos
)
