// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the hardware
// semaphore (HSEM) peripheral of the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7_cm4 stm32h7x7_cm7

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
