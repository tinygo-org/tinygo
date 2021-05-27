// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the HSEM peripheral
// of the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7

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
		return HSEM.HSEM_RLR0.Get() == (HSEM_HSEM_RLR0_LOCK | core)
	case 1:
		return HSEM.HSEM_RLR1.Get() == (HSEM_HSEM_RLR1_LOCK | core)
	case 2:
		return HSEM.HSEM_RLR2.Get() == (HSEM_HSEM_RLR2_LOCK | core)
	case 3:
		return HSEM.HSEM_RLR3.Get() == (HSEM_HSEM_RLR3_LOCK | core)
	case 4:
		return HSEM.HSEM_RLR4.Get() == (HSEM_HSEM_RLR4_LOCK | core)
	case 5:
		return HSEM.HSEM_RLR5.Get() == (HSEM_HSEM_RLR5_LOCK | core)
	case 6:
		return HSEM.HSEM_RLR6.Get() == (HSEM_HSEM_RLR6_LOCK | core)
	case 7:
		return HSEM.HSEM_RLR7.Get() == (HSEM_HSEM_RLR7_LOCK | core)
	case 8:
		return HSEM.HSEM_RLR8.Get() == (HSEM_HSEM_RLR8_LOCK | core)
	case 9:
		return HSEM.HSEM_RLR9.Get() == (HSEM_HSEM_RLR9_LOCK | core)
	case 10:
		return HSEM.HSEM_RLR10.Get() == (HSEM_HSEM_RLR10_LOCK | core)
	case 11:
		return HSEM.HSEM_RLR11.Get() == (HSEM_HSEM_RLR11_LOCK | core)
	case 12:
		return HSEM.HSEM_RLR12.Get() == (HSEM_HSEM_RLR12_LOCK | core)
	case 13:
		return HSEM.HSEM_RLR13.Get() == (HSEM_HSEM_RLR13_LOCK | core)
	case 14:
		return HSEM.HSEM_RLR14.Get() == (HSEM_HSEM_RLR14_LOCK | core)
	case 15:
		return HSEM.HSEM_RLR15.Get() == (HSEM_HSEM_RLR15_LOCK | core)
	case 16:
		return HSEM.HSEM_RLR16.Get() == (HSEM_HSEM_RLR16_LOCK | core)
	case 17:
		return HSEM.HSEM_RLR17.Get() == (HSEM_HSEM_RLR17_LOCK | core)
	case 18:
		return HSEM.HSEM_RLR18.Get() == (HSEM_HSEM_RLR18_LOCK | core)
	case 19:
		return HSEM.HSEM_RLR19.Get() == (HSEM_HSEM_RLR19_LOCK | core)
	case 20:
		return HSEM.HSEM_RLR20.Get() == (HSEM_HSEM_RLR20_LOCK | core)
	case 21:
		return HSEM.HSEM_RLR21.Get() == (HSEM_HSEM_RLR21_LOCK | core)
	case 22:
		return HSEM.HSEM_RLR22.Get() == (HSEM_HSEM_RLR22_LOCK | core)
	case 23:
		return HSEM.HSEM_RLR23.Get() == (HSEM_HSEM_RLR23_LOCK | core)
	case 24:
		return HSEM.HSEM_RLR24.Get() == (HSEM_HSEM_RLR24_LOCK | core)
	case 25:
		return HSEM.HSEM_RLR25.Get() == (HSEM_HSEM_RLR25_LOCK | core)
	case 26:
		return HSEM.HSEM_RLR26.Get() == (HSEM_HSEM_RLR26_LOCK | core)
	case 27:
		return HSEM.HSEM_RLR27.Get() == (HSEM_HSEM_RLR27_LOCK | core)
	case 28:
		return HSEM.HSEM_RLR28.Get() == (HSEM_HSEM_RLR28_LOCK | core)
	case 29:
		return HSEM.HSEM_RLR29.Get() == (HSEM_HSEM_RLR29_LOCK | core)
	case 30:
		return HSEM.HSEM_RLR30.Get() == (HSEM_HSEM_RLR30_LOCK | core)
	case 31:
		return HSEM.HSEM_RLR31.Get() == (HSEM_HSEM_RLR31_LOCK | core)
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
		HSEM.HSEM_R0.Set(core)
	case 1:
		HSEM.HSEM_R1.Set(core)
	case 2:
		HSEM.HSEM_R2.Set(core)
	case 3:
		HSEM.HSEM_R3.Set(core)
	case 4:
		HSEM.HSEM_R4.Set(core)
	case 5:
		HSEM.HSEM_R5.Set(core)
	case 6:
		HSEM.HSEM_R6.Set(core)
	case 7:
		HSEM.HSEM_R7.Set(core)
	case 8:
		HSEM.HSEM_R8.Set(core)
	case 9:
		HSEM.HSEM_R9.Set(core)
	case 10:
		HSEM.HSEM_R10.Set(core)
	case 11:
		HSEM.HSEM_R11.Set(core)
	case 12:
		HSEM.HSEM_R12.Set(core)
	case 13:
		HSEM.HSEM_R13.Set(core)
	case 14:
		HSEM.HSEM_R14.Set(core)
	case 15:
		HSEM.HSEM_R15.Set(core)
	case 16:
		HSEM.HSEM_R16.Set(core)
	case 17:
		HSEM.HSEM_R17.Set(core)
	case 18:
		HSEM.HSEM_R18.Set(core)
	case 19:
		HSEM.HSEM_R19.Set(core)
	case 20:
		HSEM.HSEM_R20.Set(core)
	case 21:
		HSEM.HSEM_R21.Set(core)
	case 22:
		HSEM.HSEM_R22.Set(core)
	case 23:
		HSEM.HSEM_R23.Set(core)
	case 24:
		HSEM.HSEM_R24.Set(core)
	case 25:
		HSEM.HSEM_R25.Set(core)
	case 26:
		HSEM.HSEM_R26.Set(core)
	case 27:
		HSEM.HSEM_R27.Set(core)
	case 28:
		HSEM.HSEM_R28.Set(core)
	case 29:
		HSEM.HSEM_R29.Set(core)
	case 30:
		HSEM.HSEM_R30.Set(core)
	case 31:
		HSEM.HSEM_R31.Set(core)
	}
}
