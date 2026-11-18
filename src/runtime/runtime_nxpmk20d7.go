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

//go:build nxp && mk20d7

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
)

const (
	watchdogUnlockSequence1 = 0xC520
	watchdogUnlockSequence2 = 0xD928
)

//go:export Reset_Handler
func main() {
	initSystem()
	arm.Asm("CPSIE i")
	initSleepTimer()

	run()
	exit(0)
}

func initSystem() {
	// from: ResetHandler

	nxp.WDOG.UNLOCK.Set(watchdogUnlockSequence1)
	nxp.WDOG.UNLOCK.Set(watchdogUnlockSequence2)
	arm.Asm("nop")
	arm.Asm("nop")
	nxp.WDOG.STCTRLH.Set(nxp.WDOG_STCTRLH_ALLOWUPDATE)

	// enable clocks to always-used peripherals
	nxp.SIM.SCGC3.Set(nxp.SIM_SCGC3_ADC1 | nxp.SIM_SCGC3_FTM2)
	nxp.SIM.SCGC5.Set(0x00043F82) // clocks active to all GPIO
	nxp.SIM.SCGC6.Set(nxp.SIM_SCGC6_RTC | nxp.SIM_SCGC6_FTM0 | nxp.SIM_SCGC6_FTM1 | nxp.SIM_SCGC6_ADC0 | nxp.SIM_SCGC6_FTFL)

	// release I/O pins hold, if we woke up from VLLS mode
	if nxp.PMC.REGSC.HasBits(nxp.PMC_REGSC_ACKISO) {
		nxp.PMC.REGSC.SetBits(nxp.PMC_REGSC_ACKISO)
	}

	// since this is a write once register, make it visible to all F_CPU's
	// so we can into other sleep modes in the future at any speed
	nxp.SMC.PMPROT.Set(nxp.SMC_PMPROT_AVLP | nxp.SMC_PMPROT_ALLS | nxp.SMC_PMPROT_AVLLS)

	preinit()

	for i := uint32(0); i <= nxp.IRQ_max; i++ {
		arm.SetPriority(i, 128)
	}

	// hardware always starts in FEI mode
	// enable capacitors for crystal
	nxp.OSC.CR.Set(nxp.OSC_CR_SC8P | nxp.OSC_CR_SC2P | nxp.OSC_CR_ERCLKEN)
	// enable osc, 8-32 MHz range, low power mode
	nxp.MCG.C2.Set(uint8((2 << nxp.MCG_C2_RANGE0_Pos) | nxp.MCG_C2_EREFS0))
	// switch to crystal as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set(uint8((2 << nxp.MCG_C1_CLKS_Pos) | (4 << nxp.MCG_C1_FRDIV_Pos)))
	// wait for crystal oscillator to begin
	for !nxp.MCG.S.HasBits(nxp.MCG_S_OSCINIT0) {
	}
	// wait for FLL to use oscillator
	for nxp.MCG.S.HasBits(nxp.MCG_S_IREFST) {
	}
	// wait for MCGOUT to use oscillator
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != (2 << nxp.MCG_S_CLKST_Pos) {
	}

	// now in FBE mode, configure PLL for 72 MHz: 16 MHz / 6 * 27
	nxp.MCG.C5.Set(5 << nxp.MCG_C5_PRDIV0_Pos)
	nxp.MCG.C6.Set(nxp.MCG_C6_PLLS | (3 << nxp.MCG_C6_VDIV0_Pos))

	// wait for PLL to start using xtal as its input
	for !nxp.MCG.S.HasBits(nxp.MCG_S_PLLST) {
	}
	// wait for PLL to lock
	for !nxp.MCG.S.HasBits(nxp.MCG_S_LOCK0) {
	}
	// now we're in PBE mode

	// config divisors: 72 MHz core, 36 MHz bus, 24 MHz flash
	nxp.SIM.CLKDIV1.Set((0 << nxp.SIM_CLKDIV1_OUTDIV1_Pos) | (1 << nxp.SIM_CLKDIV1_OUTDIV2_Pos) | (2 << nxp.SIM_CLKDIV1_OUTDIV4_Pos))

	// switch to PLL as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set((0 << nxp.MCG_C1_CLKS_Pos) | (4 << nxp.MCG_C1_FRDIV_Pos))
	// wait for PLL clock to be used
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != (3 << nxp.MCG_S_CLKST_Pos) {
	}
	// now we're in PEE mode
	// trace is CPU clock, CLKOUT=OSCERCLK0
	nxp.SIM.SOPT2.Set(nxp.SIM_SOPT2_USBSRC | nxp.SIM_SOPT2_PLLFLLSEL | nxp.SIM_SOPT2_TRACECLKSEL | (6 << nxp.SIM_SOPT2_CLKOUTSEL_Pos))

	// If the RTC oscillator isn't enabled, get it started.
	if !nxp.RTC.CR.HasBits(nxp.RTC_CR_OSCE) {
		nxp.RTC.SR.Set(0)
		nxp.RTC.CR.Set(nxp.RTC_CR_SC16P | nxp.RTC_CR_SC4P | nxp.RTC_CR_OSCE)
	}

	// initialize the SysTick counter
	initSysTick()
}

func putchar(c byte) {
	machine.PutcharUART(machine.UART0, c)
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func exit(code int) {
	abort()
}

func abort() {
	println("!!! ABORT !!!")

	m := arm.DisableInterrupts()
	arm.Asm("mov r12, #1")
	arm.Asm("msr basepri, r12")                                           // only execute interrupts of priority 0
	nxp.SystemControl.SHPR3.ClearBits(nxp.SystemControl_SHPR3_PRI_15_Msk) // set systick to priority 0
	arm.EnableInterrupts(m)

	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	var v bool
	for {
		machine.LED.Set(v)
		v = !v

		t := millisSinceBoot()
		for millisSinceBoot()-t < 60 {
			arm.Asm("wfi")
		}

		// keep polling some communication while in fault
		// mode, so we don't completely die.
		machine.PollUART(machine.UART0)
		machine.PollUART(machine.UART1)
		machine.PollUART(machine.UART2)
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}
