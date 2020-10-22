// +build cortexm0

package arm

import "runtime/volatile"

// System Control Block (SCB)
//
// SCB_Type provides the definitions for the System Control Block Registers.
type SCB_Type struct {
	CPUID volatile.Register32 // 00 CPUID Base Register
	ICSR  volatile.Register32 // 04 Interrupt Control and State Register
	_     volatile.Register32 // 08
	AIRCR volatile.Register32 // 0C Application Interrupt and Reset Control Register
	SCR   volatile.Register32 // 10 System Control Register
	CCR   volatile.Register32 // 14 Configuration Control Register
	_     volatile.Register32 // 18
	SHPR2 volatile.Register32 // 1C System Handlers Priority Registers 2
	SHPR3 volatile.Register32 // 20 System Handlers Priority Registers 3
}
