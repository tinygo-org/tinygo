// CMSIS abstraction functions.
//
// Original copyright:
//
//	Copyright (c) 2009 - 2015 ARM LIMITED
//
//	All rights reserved.
//	Redistribution and use in source and binary forms, with or without
//	modification, are permitted provided that the following conditions are met:
//	- Redistributions of source code must retain the above copyright
//	  notice, this list of conditions and the following disclaimer.
//	- Redistributions in binary form must reproduce the above copyright
//	  notice, this list of conditions and the following disclaimer in the
//	  documentation and/or other materials provided with the distribution.
//	- Neither the name of ARM nor the names of its contributors may be used
//	  to endorse or promote products derived from this software without
//	  specific prior written permission.
//
//	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
//	AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
//	IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
//	ARE DISCLAIMED. IN NO EVENT SHALL COPYRIGHT HOLDERS AND CONTRIBUTORS BE
//	LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
//	CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
//	SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
//	INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
//	CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
//	ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
//	POSSIBILITY OF SUCH DAMAGE.
package arm

import "C"
import (
	"errors"
	"runtime/volatile"
	"unsafe"
)

var errCycleCountTooLarge = errors.New("requested cycle count is too large, overflows 24 bit counter")

// Run the given assembly code. The code will be marked as having side effects,
// as it doesn't produce output and thus would normally be eliminated by the
// optimizer.
func Asm(asm string)

// Run the given inline assembly. The code will be marked as having side
// effects, as it would otherwise be optimized away. The inline assembly string
// recognizes template values in the form {name}, like so:
//
//	arm.AsmFull(
//	    "str {value}, {result}",
//	    map[string]interface{}{
//	        "value":  1
//	        "result": &dest,
//	    })
//
// You can use {} in the asm string (which expands to a register) to set the
// return value.
func AsmFull(asm string, regs map[string]interface{}) uintptr

// Run the following system call (SVCall) with 0 arguments.
func SVCall0(num uintptr) uintptr

// Run the following system call (SVCall) with 1 argument.
func SVCall1(num uintptr, a1 interface{}) uintptr

// Run the following system call (SVCall) with 2 arguments.
func SVCall2(num uintptr, a1, a2 interface{}) uintptr

// Run the following system call (SVCall) with 3 arguments.
func SVCall3(num uintptr, a1, a2, a3 interface{}) uintptr

// Run the following system call (SVCall) with 4 arguments.
func SVCall4(num uintptr, a1, a2, a3, a4 interface{}) uintptr

const (
	SCS_BASE  = 0xE000E000
	SYST_BASE = SCS_BASE + 0x0010
	NVIC_BASE = SCS_BASE + 0x0100
)

// Nested Vectored Interrupt Controller (NVIC).
//
// Source:
// http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.dui0553a/CIHIGCIF.html
type NVIC_Type struct {
	ISER [8]volatile.Register32 // Interrupt Set-enable Registers
	_    [24]uint32
	ICER [8]volatile.Register32 // Interrupt Clear-enable Registers
	_    [24]uint32
	ISPR [8]volatile.Register32 // Interrupt Set-pending Registers
	_    [24]uint32
	ICPR [8]volatile.Register32 // Interrupt Clear-pending Registers
	_    [24]uint32
	IABR [8]volatile.Register32 // Interrupt Active Bit Registers
	_    [56]uint32
	IPR  [60]volatile.Register32 // Interrupt Priority Registers
}

var NVIC = (*NVIC_Type)(unsafe.Pointer(uintptr(NVIC_BASE)))

// System Timer (SYST)
//
// Source: https://static.docs.arm.com/ddi0403/e/DDI0403E_d_armv7m_arm.pdf B3.3
type SYST_Type struct {
	SYST_CSR   volatile.Register32
	SYST_RVR   volatile.Register32
	SYST_CVR   volatile.Register32
	SYST_CALIB volatile.Register32
}

var SYST = (*SYST_Type)(unsafe.Pointer(uintptr(SYST_BASE)))

// Bitfields for SYST: System Timer
const (
	// SYST.SYST_CSR: SysTick Control and Status Register
	SYST_CSR_ENABLE_Pos    = 0x0     // Position of ENABLE field.
	SYST_CSR_ENABLE_Msk    = 0x1     // Bit mask of ENABLE field.
	SYST_CSR_ENABLE        = 0x1     // Bit ENABLE.
	SYST_CSR_TICKINT_Pos   = 0x1     // Position of TICKINT field.
	SYST_CSR_TICKINT_Msk   = 0x2     // Bit mask of TICKINT field.
	SYST_CSR_TICKINT       = 0x2     // Bit TICKINT.
	SYST_CSR_CLKSOURCE_Pos = 0x2     // Position of CLKSOURCE field.
	SYST_CSR_CLKSOURCE_Msk = 0x4     // Bit mask of CLKSOURCE field.
	SYST_CSR_CLKSOURCE     = 0x4     // Bit CLKSOURCE.
	SYST_CSR_COUNTFLAG_Pos = 0x10    // Position of COUNTFLAG field.
	SYST_CSR_COUNTFLAG_Msk = 0x10000 // Bit mask of COUNTFLAG field.
	SYST_CSR_COUNTFLAG     = 0x10000 // Bit COUNTFLAG.

	// SYST.SYST_RVR: SysTick Reload Value Register
	SYST_RVR_RELOAD_Pos = 0x0      // Position of RELOAD field.
	SYST_RVR_RELOAD_Msk = 0xffffff // Bit mask of RELOAD field.

	// SYST.SYST_CVR: SysTick Current Value Register
	SYST_CVR_CURRENT_Pos = 0x0      // Position of CURRENT field.
	SYST_CVR_CURRENT_Msk = 0xffffff // Bit mask of CURRENT field.

	// SYST.SYST_CALIB: SysTick Calibration Value Register
	SYST_CALIB_TENMS_Pos = 0x0        // Position of TENMS field.
	SYST_CALIB_TENMS_Msk = 0xffffff   // Bit mask of TENMS field.
	SYST_CALIB_SKEW_Pos  = 0x1e       // Position of SKEW field.
	SYST_CALIB_SKEW_Msk  = 0x40000000 // Bit mask of SKEW field.
	SYST_CALIB_SKEW      = 0x40000000 // Bit SKEW.
	SYST_CALIB_NOREF_Pos = 0x1f       // Position of NOREF field.
	SYST_CALIB_NOREF_Msk = 0x80000000 // Bit mask of NOREF field.
	SYST_CALIB_NOREF     = 0x80000000 // Bit NOREF.
)

// Enable the given interrupt number.
func EnableIRQ(irq uint32) {
	NVIC.ISER[irq>>5].Set(1 << (irq & 0x1F))
}

// Disable the given interrupt number.
func DisableIRQ(irq uint32) {
	NVIC.ICER[irq>>5].Set(1 << (irq & 0x1F))
}

// Set the priority of the given interrupt number.
// Note that the priority is given as a 0-255 number, where some of the lower
// bits are not implemented by the hardware. For example, to set a low interrupt
// priority, use 0xc0, which is equivalent to using priority level 5 when the
// hardware has 8 priority levels. Also note that the priority level is inverted
// in ARM: a lower number means it is a more important interrupt and will
// interrupt ISRs with a higher interrupt priority.
func SetPriority(irq uint32, priority uint32) {
	// Details:
	// http://infocenter.arm.com/help/index.jsp?topic=/com.arm.doc.dui0553a/Cihgjeed.html
	regnum := irq / 4
	regpos := irq % 4
	mask := uint32(0xff) << (regpos * 8) // bits to clear
	priority = priority << (regpos * 8)  // bits to set
	NVIC.IPR[regnum].Set((uint32(NVIC.IPR[regnum].Get()) &^ mask) | priority)
}

// DisableInterrupts disables all interrupts, and returns the old interrupt
// state.
//
//export DisableInterrupts
func DisableInterrupts() uintptr

// EnableInterrupts enables all interrupts again. The value passed in must be
// the mask returned by DisableInterrupts.
//
//export EnableInterrupts
func EnableInterrupts(mask uintptr)

// Set up the system timer to generate periodic tick events.
// This will cause SysTick_Handler to fire once per tick.
// The cyclecount parameter is a counter value which can range from 0 to
// 0xffffff.  A value of 0 disables the timer.
func SetupSystemTimer(cyclecount uint32) error {
	// turn it off
	SYST.SYST_CSR.ClearBits(SYST_CSR_TICKINT | SYST_CSR_ENABLE)
	if cyclecount == 0 {
		// leave the system timer turned off.
		return nil
	}
	if cyclecount&SYST_RVR_RELOAD_Msk != cyclecount {
		// The cycle refresh register is only 24 bits wide.  The user-specified value will overflow.
		return errCycleCountTooLarge
	}

	// set refresh count
	SYST.SYST_RVR.Set(cyclecount)
	// set current counter value
	SYST.SYST_CVR.Set(cyclecount)
	// enable clock, enable SysTick interrupt when clock reaches 0, run it off of the processor clock
	SYST.SYST_CSR.SetBits(SYST_CSR_TICKINT | SYST_CSR_ENABLE | SYST_CSR_CLKSOURCE)
	return nil
}
