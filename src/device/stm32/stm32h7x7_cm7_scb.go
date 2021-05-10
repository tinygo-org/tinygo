// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the SCB peripheral of
// the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to the Cortex-M7 core only.

// +build stm32h7x7_cm7

package stm32

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

const (
	SCB_CCR_ICACHEEN_Pos = 17                          // SCB CCR: IC Position
	SCB_CCR_ICACHEEN_Msk = (1 << SCB_CCR_ICACHEEN_Pos) // SCB CCR: IC Mask
	SCB_CCR_DCACHEEN_Pos = 16                          // SCB CCR: DC Position
	SCB_CCR_DCACHEEN_Msk = (1 << SCB_CCR_DCACHEEN_Pos) // SCB CCR: DC Mask

	SCB_DCISW_WAY_Pos = 30                         // SCB DCISW: WAY Position
	SCB_DCISW_WAY_Msk = 3 << SCB_DCISW_WAY_Pos     // SCB DCISW: WAY Mask
	SCB_DCISW_SET_Pos = 5                          // SCB DCISW: SET Position
	SCB_DCISW_SET_Msk = 0x1FF << SCB_DCISW_SET_Pos // SCB DCISW: SET Mask

	SCB_DCCISW_WAY_Pos     = 30                          // SCB DCCISW: WAY Position
	SCB_DCCISW_WAY_Msk     = 3 << SCB_DCCISW_WAY_Pos     // SCB DCCISW: WAY Mask
	SCB_DCCISW_SET_Pos     = 5                           // SCB DCCISW: SET Position
	SCB_DCCISW_SET_Msk     = 0x1FF << SCB_DCCISW_SET_Pos // SCB DCCISW: SET Mask
	SCB_DCCISW_ASSOC_Pos   = 0x3                         // SCB DCCISW: ASSOCIATIVITY Position
	SCB_DCCISW_ASSOC_Msk   = 0x1FF8                      // SCB DCCISW: ASSOCIATIVITY Mask
	SCB_DCCISW_NUMSETS_Pos = 0xD                         // SCB DCCISW: NUMSETS Position
	SCB_DCCISW_NUMSETS_Msk = 0xFFFE000                   // SCB DCCISW: NUMSETS Mask

	SCB_SHCSR_USGFAULTENA_Pos    = 18                                // SCB SHCSR: USGFAULTENA Position
	SCB_SHCSR_USGFAULTENA_Msk    = 1 << SCB_SHCSR_USGFAULTENA_Pos    // SCB SHCSR: USGFAULTENA Mask
	SCB_SHCSR_BUSFAULTENA_Pos    = 17                                // SCB SHCSR: BUSFAULTENA Position
	SCB_SHCSR_BUSFAULTENA_Msk    = 1 << SCB_SHCSR_BUSFAULTENA_Pos    // SCB SHCSR: BUSFAULTENA Mask
	SCB_SHCSR_MEMFAULTENA_Pos    = 16                                // SCB SHCSR: MEMFAULTENA Position
	SCB_SHCSR_MEMFAULTENA_Msk    = 1 << SCB_SHCSR_MEMFAULTENA_Pos    // SCB SHCSR: MEMFAULTENA Mask
	SCB_SHCSR_SVCALLPENDED_Pos   = 15                                // SCB SHCSR: SVCALLPENDED Position
	SCB_SHCSR_SVCALLPENDED_Msk   = 1 << SCB_SHCSR_SVCALLPENDED_Pos   // SCB SHCSR: SVCALLPENDED Mask
	SCB_SHCSR_BUSFAULTPENDED_Pos = 14                                // SCB SHCSR: BUSFAULTPENDED Position
	SCB_SHCSR_BUSFAULTPENDED_Msk = 1 << SCB_SHCSR_BUSFAULTPENDED_Pos // SCB SHCSR: BUSFAULTPENDED Mask
	SCB_SHCSR_MEMFAULTPENDED_Pos = 13                                // SCB SHCSR: MEMFAULTPENDED Position
	SCB_SHCSR_MEMFAULTPENDED_Msk = 1 << SCB_SHCSR_MEMFAULTPENDED_Pos // SCB SHCSR: MEMFAULTPENDED Mask
	SCB_SHCSR_USGFAULTPENDED_Pos = 12                                // SCB SHCSR: USGFAULTPENDED Position
	SCB_SHCSR_USGFAULTPENDED_Msk = 1 << SCB_SHCSR_USGFAULTPENDED_Pos // SCB SHCSR: USGFAULTPENDED Mask
	SCB_SHCSR_SYSTICKACT_Pos     = 11                                // SCB SHCSR: SYSTICKACT Position
	SCB_SHCSR_SYSTICKACT_Msk     = 1 << SCB_SHCSR_SYSTICKACT_Pos     // SCB SHCSR: SYSTICKACT Mask
	SCB_SHCSR_PENDSVACT_Pos      = 10                                // SCB SHCSR: PENDSVACT Position
	SCB_SHCSR_PENDSVACT_Msk      = 1 << SCB_SHCSR_PENDSVACT_Pos      // SCB SHCSR: PENDSVACT Mask
	SCB_SHCSR_MONITORACT_Pos     = 8                                 // SCB SHCSR: MONITORACT Position
	SCB_SHCSR_MONITORACT_Msk     = 1 << SCB_SHCSR_MONITORACT_Pos     // SCB SHCSR: MONITORACT Mask
	SCB_SHCSR_SVCALLACT_Pos      = 7                                 // SCB SHCSR: SVCALLACT Position
	SCB_SHCSR_SVCALLACT_Msk      = 1 << SCB_SHCSR_SVCALLACT_Pos      // SCB SHCSR: SVCALLACT Mask
	SCB_SHCSR_USGFAULTACT_Pos    = 3                                 // SCB SHCSR: USGFAULTACT Position
	SCB_SHCSR_USGFAULTACT_Msk    = 1 << SCB_SHCSR_USGFAULTACT_Pos    // SCB SHCSR: USGFAULTACT Mask
	SCB_SHCSR_BUSFAULTACT_Pos    = 1                                 // SCB SHCSR: BUSFAULTACT Position
	SCB_SHCSR_BUSFAULTACT_Msk    = 1 << SCB_SHCSR_BUSFAULTACT_Pos    // SCB SHCSR: BUSFAULTACT Mask
	SCB_SHCSR_MEMFAULTACT_Pos    = 0                                 // SCB SHCSR: MEMFAULTACT Position
	SCB_SHCSR_MEMFAULTACT_Msk    = 1 << SCB_SHCSR_MEMFAULTACT_Pos    // SCB SHCSR: MEMFAULTACT Mask
)

var (
	// Offset: 0x080 (R/ )  Cache Size ID Register
	SCB_CCSIDR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x080)))
	// Offset: 0x084 (R/W)  Cache Size Selection Register
	SCB_CSSELR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x084)))
	// Offset: 0x250 ( /W)  I-Cache Invalidate All to PoU
	SCB_ICIALLU = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x250)))
	// Offset: 0x260 ( /W)  D-Cache Invalidate by Set-way
	SCB_DCISW = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x260)))
	// Offset: 0x274 ( /W)  D-Cache Clean and Invalidate by Set-way
	SCB_DCCISW = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x274)))
	// Offset: 0x024 (R/W)  System Handler Control and State Register
	SCB_SHCSR = (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(SCB)) + 0x024)))
)

func (scb *SCB_Type) EnableICache(enable bool) {
	if enable == scb.CCR.HasBits(SCB_CCR_ICACHEEN_Msk) {
		return
	}
	if enable {
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
		SCB_ICIALLU.Set(0)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
		scb.CCR.SetBits(SCB_CCR_ICACHEEN_Msk)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
	} else {
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
		scb.CCR.ClearBits(SCB_CCR_ICACHEEN_Msk)
		SCB_ICIALLU.Set(0)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
	}
}

var (
	scbDcci volatile.Register32
	scbSets volatile.Register32
	scbWays volatile.Register32
)

func (scb *SCB_Type) EnableDCache(enable bool) {
	if enable == scb.CCR.HasBits(SCB_CCR_DCACHEEN_Msk) {
		return
	}
	if enable {
		SCB_CSSELR.Set(0)
		arm.AsmFull(`
			dsb 0xF
		`, nil)
		scbDcci.Set(SCB_CCSIDR.Get())
		scbSets.Set((scbDcci.Get() & SCB_DCCISW_NUMSETS_Msk) >> SCB_DCCISW_NUMSETS_Pos)
		for scbSets.Get() != 0 {
			scbWays.Set((scbDcci.Get() & SCB_DCCISW_ASSOC_Msk) >> SCB_DCCISW_ASSOC_Pos)
			for scbWays.Get() != 0 {
				SCB_DCISW.Set(((scbSets.Get() << SCB_DCISW_SET_Pos) & SCB_DCISW_SET_Msk) |
					((scbWays.Get() << SCB_DCISW_WAY_Pos) & SCB_DCISW_WAY_Msk))
				scbWays.Set(scbWays.Get() - 1)
			}
			scbSets.Set(scbSets.Get() - 1)
		}
		arm.AsmFull(`
			dsb 0xF
		`, nil)
		scb.CCR.SetBits(SCB_CCR_DCACHEEN_Msk)
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
	} else {
		var ()
		SCB_CSSELR.Set(0)
		arm.AsmFull(`
			dsb 0xF
		`, nil)
		scb.CCR.ClearBits(SCB_CCR_DCACHEEN_Msk)
		arm.AsmFull(`
			dsb 0xF
		`, nil)
		scbDcci.Set(SCB_CCSIDR.Get())
		scbSets.Set((scbDcci.Get() & SCB_DCCISW_NUMSETS_Msk) >> SCB_DCCISW_NUMSETS_Pos)
		for scbSets.Get() != 0 {
			scbWays.Set((scbDcci.Get() & SCB_DCCISW_ASSOC_Msk) >> SCB_DCCISW_ASSOC_Pos)
			for scbWays.Get() != 0 {
				SCB_DCCISW.Set(((scbSets.Get() << SCB_DCCISW_SET_Pos) & SCB_DCCISW_SET_Msk) |
					((scbWays.Get() << SCB_DCCISW_WAY_Pos) & SCB_DCCISW_WAY_Msk))
				scbWays.Set(scbWays.Get() - 1)
			}
			scbSets.Set(scbSets.Get() - 1)
		}
		arm.AsmFull(`
			dsb 0xF
			isb 0xF
		`, nil)
	}
}
