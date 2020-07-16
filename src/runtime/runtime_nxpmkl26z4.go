// Derivative work of Teensyduino Core Library
// http://www.pjrc.com/teensy/
// Copyright (c) 2017 PJRC.COM, LLC.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// 1. The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// 2. If the Software is incorporated into a build system that allows
// selection among a list of target devices, then similar target
// devices manufactured by PJRC.COM must be included in the list of
// target devices and selectable in the same manner.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
// BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// +build nxp,mkl26z4

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
)

const asyncScheduler = false

const (
	_DEFAULT_FTM_MOD      = 49152 - 1
	_DEFAULT_FTM_PRESCALE = 1
)

//go:export Reset_Handler
func main() {
	initSystem()
	initSysTick()
	arm.Asm("CPSIE i")
	initInternal()

	run()
	abort()
}

func initSystem() {
	// from: ResetHandler

	// enable clocks to always-used peripherals
	nxp.SIM.SCGC4.Set(nxp.SIM_SCGC4_USBOTG | 0xF0000030)
	nxp.SIM.SCGC5.Set(0x00003F82) // clocks active to all GPIO
	nxp.SIM.SCGC6.Set(nxp.SIM_SCGC6_ADC0 | nxp.SIM_SCGC6_TPM0 | nxp.SIM_SCGC6_TPM1 | nxp.SIM_SCGC6_TPM2 | nxp.SIM_SCGC6_FTF)

	// release I/O pins hold, if we woke up from VLLS mode
	if nxp.PMC.REGSC.HasBits(nxp.PMC_REGSC_ACKISO) {
		nxp.PMC.REGSC.SetBits(nxp.PMC_REGSC_ACKISO)
	}

	// since this is a write once register, make it visible to all F_CPU's
	// so we can into other sleep modes in the future at any speed
	nxp.SMC.PMPROT.Set(nxp.SMC_PMPROT_AVLP | nxp.SMC_PMPROT_ALLS | nxp.SMC_PMPROT_AVLLS)

	preinit()

	// default all interrupts to medium priority level
	for i := uint32(0); i <= nxp.IRQ_max; i++ {
		arm.SetPriority(i, 128)
	}

	// now program the clock dividers
	// config divisors: 48 MHz core, 48 MHz bus, 24 MHz flash, USB = 96 / 2
	nxp.SIM.CLKDIV1.Set((1 << nxp.SIM_CLKDIV1_OUTDIV1_Pos) | (1 << nxp.SIM_CLKDIV1_OUTDIV4_Pos))

	// switch to PLL as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set((0 << nxp.MCG_C1_CLKS_Pos) | (4 << nxp.MCG_C1_FRDIV_Pos))
	// wait for PLL clock to be used
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != (3 << nxp.MCG_S_CLKST_Pos) {
	}
	// now we're in PEE mode
	// trace is CPU clock, CLKOUT=OSCERCLK0
	nxp.SIM.SOPT2.Set(nxp.SIM_SOPT2_USBSRC | nxp.SIM_SOPT2_PLLFLLSEL | (6 << nxp.SIM_SOPT2_CLKOUTSEL_Pos) | (1 << nxp.SIM_SOPT2_UART0SRC_Pos) | (1 << nxp.SIM_SOPT2_TPMSRC_Pos))
}

func initInternal() {
	// from: _init_Teensyduino_internal_

	// arm.EnableIRQ(nxp.IRQ_PORTA)
	// arm.EnableIRQ(nxp.IRQ_PORTC_PORTD)

	nxp.TPM0.CNT.Set(0)
	nxp.TPM0.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.TPM0.C0SC.Set(0x28) // MSnB:MSnA = 10, ELSnB:ELSnA = 10
	nxp.TPM0.C1SC.Set(0x28)
	nxp.TPM0.C2SC.Set(0x28)
	nxp.TPM0.C3SC.Set(0x28)
	nxp.TPM0.C4SC.Set(0x28)
	nxp.TPM0.C5SC.Set(0x28)
	nxp.TPM0.SC.Set((1 << nxp.TPM_SC_CMOD_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.TPM_SC_PS_Pos))

	nxp.TPM1.CNT.Set(0)
	nxp.TPM1.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.TPM1.C0SC.Set(0x28)
	nxp.TPM1.C1SC.Set(0x28)
	nxp.TPM1.SC.Set((1 << nxp.TPM_SC_CMOD_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.TPM_SC_PS_Pos))

	nxp.TPM2.CNT.Set(0)
	nxp.TPM2.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.TPM2.C0SC.Set(0x28)
	nxp.TPM2.C1SC.Set(0x28)
	nxp.TPM2.SC.Set((1 << nxp.TPM_SC_CMOD_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.TPM_SC_PS_Pos))
}

func postinit() {}

func putchar(c byte) {
	machine.PutcharUART(&machine.UART0, c)
}

func abort() {
	println("!!! ABORT !!!")

	// disable all interrupts
	arm.DisableInterrupts()

	// lock up forever
	for {
		arm.Asm("wfi")
	}
}
