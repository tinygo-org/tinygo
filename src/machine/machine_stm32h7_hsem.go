//go:build stm32 && stm32h753

package machine

import (
	"device/stm32"
	"runtime/volatile"
	"unsafe"
)

var (
	HSEM_CORE1 = (*HSEM_CORE_Type)(unsafe.Pointer((uintptr(unsafe.Pointer(stm32.HSEM)) + 0x100)))
	HSEM_CORE2 = (*HSEM_CORE_Type)(unsafe.Pointer((uintptr(unsafe.Pointer(stm32.HSEM)) + 0x110)))
)

// HSEM_CORE represents the registers for a core's hardware semaphore interrupts.
type HSEM_CORE_Type struct {
	IER  volatile.Register32 // HSEM Interrupt enable register         Address offset: HSEM + 0x100 + n*0x10 (Interrupt N (0 or 1))
	ICR  volatile.Register32 // HSEM Interrupt clear register          Address offset: HSEM + 0x104 + n*0x10 (Interrupt N (0 or 1))
	ISR  volatile.Register32 // HSEM Interrupt Status register         Address offset: HSEM + 0x108 + n*0x10 (Interrupt N (0 or 1))
	MISR volatile.Register32 // HSEM Interrupt Masked Status register  Address offset: HSEM + 0x10C + n*0x10 (Interrupt N (0 or 1))
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
		return stm32.HSEM.RLR0.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 1:
		return stm32.HSEM.RLR1.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 2:
		return stm32.HSEM.RLR2.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 3:
		return stm32.HSEM.RLR3.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 4:
		return stm32.HSEM.RLR4.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 5:
		return stm32.HSEM.RLR5.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 6:
		return stm32.HSEM.RLR6.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 7:
		return stm32.HSEM.RLR7.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 8:
		return stm32.HSEM.RLR8.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 9:
		return stm32.HSEM.RLR9.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 10:
		return stm32.HSEM.RLR10.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 11:
		return stm32.HSEM.RLR11.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 12:
		return stm32.HSEM.RLR12.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 13:
		return stm32.HSEM.RLR13.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 14:
		return stm32.HSEM.RLR14.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 15:
		return stm32.HSEM.RLR15.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 16:
		return stm32.HSEM.RLR16.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 17:
		return stm32.HSEM.RLR17.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 18:
		return stm32.HSEM.RLR18.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 19:
		return stm32.HSEM.RLR19.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 20:
		return stm32.HSEM.RLR20.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 21:
		return stm32.HSEM.RLR21.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 22:
		return stm32.HSEM.RLR22.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 23:
		return stm32.HSEM.RLR23.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 24:
		return stm32.HSEM.RLR24.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 25:
		return stm32.HSEM.RLR25.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 26:
		return stm32.HSEM.RLR26.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 27:
		return stm32.HSEM.RLR27.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 28:
		return stm32.HSEM.RLR28.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 29:
		return stm32.HSEM.RLR29.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 30:
		return stm32.HSEM.RLR30.Get() == (stm32.HSEM_RLR_LOCK | core)
	case 31:
		return stm32.HSEM.RLR31.Get() == (stm32.HSEM_RLR_LOCK | core)
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
		stm32.HSEM.R0.Set(core)
	case 1:
		stm32.HSEM.R1.Set(core)
	case 2:
		stm32.HSEM.R2.Set(core)
	case 3:
		stm32.HSEM.R3.Set(core)
	case 4:
		stm32.HSEM.R4.Set(core)
	case 5:
		stm32.HSEM.R5.Set(core)
	case 6:
		stm32.HSEM.R6.Set(core)
	case 7:
		stm32.HSEM.R7.Set(core)
	case 8:
		stm32.HSEM.R8.Set(core)
	case 9:
		stm32.HSEM.R9.Set(core)
	case 10:
		stm32.HSEM.R10.Set(core)
	case 11:
		stm32.HSEM.R11.Set(core)
	case 12:
		stm32.HSEM.R12.Set(core)
	case 13:
		stm32.HSEM.R13.Set(core)
	case 14:
		stm32.HSEM.R14.Set(core)
	case 15:
		stm32.HSEM.R15.Set(core)
	case 16:
		stm32.HSEM.R16.Set(core)
	case 17:
		stm32.HSEM.R17.Set(core)
	case 18:
		stm32.HSEM.R18.Set(core)
	case 19:
		stm32.HSEM.R19.Set(core)
	case 20:
		stm32.HSEM.R20.Set(core)
	case 21:
		stm32.HSEM.R21.Set(core)
	case 22:
		stm32.HSEM.R22.Set(core)
	case 23:
		stm32.HSEM.R23.Set(core)
	case 24:
		stm32.HSEM.R24.Set(core)
	case 25:
		stm32.HSEM.R25.Set(core)
	case 26:
		stm32.HSEM.R26.Set(core)
	case 27:
		stm32.HSEM.R27.Set(core)
	case 28:
		stm32.HSEM.R28.Set(core)
	case 29:
		stm32.HSEM.R29.Set(core)
	case 30:
		stm32.HSEM.R30.Set(core)
	case 31:
		stm32.HSEM.R31.Set(core)
	}
}
