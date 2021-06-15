// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the HSEM peripheral
// of the STM32H7x5 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x5

package stm32

import (
	"runtime/volatile"
	"unsafe"
)

var (
	HSEM_CORE1 = (*HSEM_CORE_Type)(unsafe.Pointer((uintptr(unsafe.Pointer(HSEM)) + 0x100)))
	HSEM_CORE2 = (*HSEM_CORE_Type)(unsafe.Pointer((uintptr(unsafe.Pointer(HSEM)) + 0x110)))
)

// HSEM_CORE
type HSEM_CORE_Type struct {
	IER  volatile.Register32 // HSEM Interrupt enable register         Address offset: HSEM + 1N0h (Interrupt N (0 or 1))
	ICR  volatile.Register32 // HSEM Interrupt clear register          Address offset: HSEM + 1N4h (Interrupt N (0 or 1))
	ISR  volatile.Register32 // HSEM Interrupt Status register         Address offset: HSEM + 1N8h (Interrupt N (0 or 1))
	MISR volatile.Register32 // HSEM Interrupt Masked Status register  Address offset: HSEM + 1NCh (Interrupt N (0 or 1))
}

type HSEM_ID_Type uint8

// Lock performs a 1-step (read) lock on the receiver semaphore ID.
// Semaphores can be used to ensure synchronization between processes running on
// different cores. Lock provides a non-blocking mechanism to lock semaphores
// in an atomic way.
// Returns true if and only if the semaphore lock is acquired or the given core
// already has the semaphore locked.
func (id HSEM_ID_Type) Lock(core uint32) bool {
	switch id {
	case 0:
		return HSEM.RLR0.Get() == (HSEM_RLR0_LOCK | core)
	case 1:
		return HSEM.RLR1.Get() == (HSEM_RLR1_LOCK | core)
	case 2:
		return HSEM.RLR2.Get() == (HSEM_RLR2_LOCK | core)
	case 3:
		return HSEM.RLR3.Get() == (HSEM_RLR3_LOCK | core)
	case 4:
		return HSEM.RLR4.Get() == (HSEM_RLR4_LOCK | core)
	case 5:
		return HSEM.RLR5.Get() == (HSEM_RLR5_LOCK | core)
	case 6:
		return HSEM.RLR6.Get() == (HSEM_RLR6_LOCK | core)
	case 7:
		return HSEM.RLR7.Get() == (HSEM_RLR7_LOCK | core)
	case 8:
		return HSEM.RLR8.Get() == (HSEM_RLR8_LOCK | core)
	case 9:
		return HSEM.RLR9.Get() == (HSEM_RLR9_LOCK | core)
	case 10:
		return HSEM.RLR10.Get() == (HSEM_RLR10_LOCK | core)
	case 11:
		return HSEM.RLR11.Get() == (HSEM_RLR11_LOCK | core)
	case 12:
		return HSEM.RLR12.Get() == (HSEM_RLR12_LOCK | core)
	case 13:
		return HSEM.RLR13.Get() == (HSEM_RLR13_LOCK | core)
	case 14:
		return HSEM.RLR14.Get() == (HSEM_RLR14_LOCK | core)
	case 15:
		return HSEM.RLR15.Get() == (HSEM_RLR15_LOCK | core)
	case 16:
		return HSEM.RLR16.Get() == (HSEM_RLR16_LOCK | core)
	case 17:
		return HSEM.RLR17.Get() == (HSEM_RLR17_LOCK | core)
	case 18:
		return HSEM.RLR18.Get() == (HSEM_RLR18_LOCK | core)
	case 19:
		return HSEM.RLR19.Get() == (HSEM_RLR19_LOCK | core)
	case 20:
		return HSEM.RLR20.Get() == (HSEM_RLR20_LOCK | core)
	case 21:
		return HSEM.RLR21.Get() == (HSEM_RLR21_LOCK | core)
	case 22:
		return HSEM.RLR22.Get() == (HSEM_RLR22_LOCK | core)
	case 23:
		return HSEM.RLR23.Get() == (HSEM_RLR23_LOCK | core)
	case 24:
		return HSEM.RLR24.Get() == (HSEM_RLR24_LOCK | core)
	case 25:
		return HSEM.RLR25.Get() == (HSEM_RLR25_LOCK | core)
	case 26:
		return HSEM.RLR26.Get() == (HSEM_RLR26_LOCK | core)
	case 27:
		return HSEM.RLR27.Get() == (HSEM_RLR27_LOCK | core)
	case 28:
		return HSEM.RLR28.Get() == (HSEM_RLR28_LOCK | core)
	case 29:
		return HSEM.RLR29.Get() == (HSEM_RLR29_LOCK | core)
	case 30:
		return HSEM.RLR30.Get() == (HSEM_RLR30_LOCK | core)
	case 31:
		return HSEM.RLR31.Get() == (HSEM_RLR31_LOCK | core)
	}
	return false
}

// Unlock releases the lock on the receiver semaphore ID.
// Semaphores can be used to ensure synchronization between processes running on
// different cores. Unlock provides a non-blocking mechanism to unlock
// semaphores in an atomic way.
func (id HSEM_ID_Type) Unlock(core uint32) {
	switch id {
	case 0:
		HSEM.R0.Set(core)
	case 1:
		HSEM.R1.Set(core)
	case 2:
		HSEM.R2.Set(core)
	case 3:
		HSEM.R3.Set(core)
	case 4:
		HSEM.R4.Set(core)
	case 5:
		HSEM.R5.Set(core)
	case 6:
		HSEM.R6.Set(core)
	case 7:
		HSEM.R7.Set(core)
	case 8:
		HSEM.R8.Set(core)
	case 9:
		HSEM.R9.Set(core)
	case 10:
		HSEM.R10.Set(core)
	case 11:
		HSEM.R11.Set(core)
	case 12:
		HSEM.R12.Set(core)
	case 13:
		HSEM.R13.Set(core)
	case 14:
		HSEM.R14.Set(core)
	case 15:
		HSEM.R15.Set(core)
	case 16:
		HSEM.R16.Set(core)
	case 17:
		HSEM.R17.Set(core)
	case 18:
		HSEM.R18.Set(core)
	case 19:
		HSEM.R19.Set(core)
	case 20:
		HSEM.R20.Set(core)
	case 21:
		HSEM.R21.Set(core)
	case 22:
		HSEM.R22.Set(core)
	case 23:
		HSEM.R23.Set(core)
	case 24:
		HSEM.R24.Set(core)
	case 25:
		HSEM.R25.Set(core)
	case 26:
		HSEM.R26.Set(core)
	case 27:
		HSEM.R27.Set(core)
	case 28:
		HSEM.R28.Set(core)
	case 29:
		HSEM.R29.Set(core)
	case 30:
		HSEM.R30.Set(core)
	case 31:
		HSEM.R31.Set(core)
	}
}
