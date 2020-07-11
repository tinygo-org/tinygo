// +build cortexm3 cortexm4

package arm

import "runtime/volatile"

// System Control Block (SCB)
//
// SCB_Type provides the definitions for the System Control Block Registers.
type SCB_Type struct {
	CPUID volatile.Register32 // 00 CPUID Base Register
	ICSR  volatile.Register32 // 04 Interrupt Control and State Register
	VTOR  volatile.Register32 // 08 Vector Table Offset Register
	AIRCR volatile.Register32 // 0C Application Interrupt and Reset Control Register
	SCR   volatile.Register32 // 10 System Control Register
	CCR   volatile.Register32 // 14 Configuration Control Register
	SHPR1 volatile.Register32 // 18 System Handlers Priority Registers 1
	SHPR2 volatile.Register32 // 1C System Handlers Priority Registers 2
	SHPR3 volatile.Register32 // 20 System Handlers Priority Registers 3
	SHCSR volatile.Register32 // 24 System Handler Control and State Register
	CFSR  volatile.Register32 // 28 Configurable Fault Status Register
	HFSR  volatile.Register32 // 2C HardFault Status Register
	_     volatile.Register32 // 30
	MMFAR volatile.Register32 // 34 MemManage Fault Address Register
	BFAR  volatile.Register32 // 38 BusFault Address Register
	AFSR  volatile.Register32 // 3C Auxiliary Fault Status Register
}
