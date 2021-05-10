// Hand created file. DO NOT DELETE.
// Type definitions, fields, and constants associated with the external
// interrupt (EXTI) peripheral of the STM32H7x7 family of dual-core MCUs.
// These definitions are applicable to both the Cortex-M7 and Cortex-M4 cores.

// +build stm32h7x7_cm4 stm32h7x7_cm7

package stm32

import (
	"runtime/volatile"
	"unsafe"
)

var (
	EXTI_CORE1 = (*EXTI_CORE_Type)(unsafe.Pointer(uintptr(0x58000080)))
	EXTI_CORE2 = (*EXTI_CORE_Type)(unsafe.Pointer(uintptr(0x580000C0)))
)

// EXTI_CORE
type EXTI_CORE_Type struct {
	IMR1 volatile.Register32 // EXTI Interrupt mask register Address offset: 0x00
	EMR1 volatile.Register32 // EXTI Event mask register     Address offset: 0x04
	PR1  volatile.Register32 // EXTI Pending register        Address offset: 0x08
	_    [4]byte             // Reserved, 0x0C
	IMR2 volatile.Register32 // EXTI Interrupt mask register Address offset: 0x10
	EMR2 volatile.Register32 // EXTI Event mask register     Address offset: 0x14
	PR2  volatile.Register32 // EXTI Pending register        Address offset: 0x18
	_    [4]byte             // Reserved, 0x1C
	IMR3 volatile.Register32 // EXTI Interrupt mask register Address offset: 0x20
	EMR3 volatile.Register32 // EXTI Event mask register     Address offset: 0x24
	PR3  volatile.Register32 // EXTI Pending register        Address offset: 0x28
}
