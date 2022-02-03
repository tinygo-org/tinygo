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

//go:build nxp && mk66f18
// +build nxp,mk66f18

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
)

const (
	watchdogUnlockSequence1 = 0xC520
	watchdogUnlockSequence2 = 0xD928

	_DEFAULT_FTM_MOD      = 61440 - 1
	_DEFAULT_FTM_PRESCALE = 1
)

const (
	_SIM_SOPT2_IRC48SEL = 3 << nxp.SIM_SOPT2_PLLFLLSEL_Pos
	_SMC_PMCTRL_HSRUN   = 3 << nxp.SMC_PMCTRL_RUNM_Pos
	_SMC_PMSTAT_HSRUN   = 0x80 << nxp.SMC_PMSTAT_PMSTAT_Pos
)

//go:export Reset_Handler
func main() {
	initSystem()
	arm.Asm("CPSIE i")
	initInternal()

	run()
	exit(0)
}

func initSystem() {
	// from: ResetHandler

	nxp.WDOG.UNLOCK.Set(watchdogUnlockSequence1)
	nxp.WDOG.UNLOCK.Set(watchdogUnlockSequence2)
	arm.Asm("nop")
	arm.Asm("nop")
	// TODO: hook for overriding? 'startupEarlyHook'
	nxp.WDOG.STCTRLH.Set(nxp.WDOG_STCTRLH_ALLOWUPDATE)

	// enable clocks to always-used peripherals
	nxp.SIM.SCGC3.Set(nxp.SIM_SCGC3_ADC1 | nxp.SIM_SCGC3_FTM2 | nxp.SIM_SCGC3_FTM3)
	nxp.SIM.SCGC5.Set(0x00043F82) // clocks active to all GPIO
	nxp.SIM.SCGC6.Set(nxp.SIM_SCGC6_RTC | nxp.SIM_SCGC6_FTM0 | nxp.SIM_SCGC6_FTM1 | nxp.SIM_SCGC6_ADC0 | nxp.SIM_SCGC6_FTF)
	nxp.LMEM.PCCCR.Set(0x85000003)

	// release I/O pins hold, if we woke up from VLLS mode
	if nxp.PMC.REGSC.HasBits(nxp.PMC_REGSC_ACKISO) {
		nxp.PMC.REGSC.SetBits(nxp.PMC_REGSC_ACKISO)
	}

	// since this is a write once register, make it visible to all F_CPU's
	// so we can into other sleep modes in the future at any speed
	nxp.SMC.PMPROT.Set(nxp.SMC_PMPROT_AHSRUN | nxp.SMC_PMPROT_AVLP | nxp.SMC_PMPROT_ALLS | nxp.SMC_PMPROT_AVLLS)

	preinit()

	// copy the vector table to RAM default all interrupts to medium priority level
	// for (i=0; i < NVIC_NUM_INTERRUPTS + 16; i++) _VectorsRam[i] = _VectorsFlash[i];
	for i := uint32(0); i <= nxp.IRQ_max; i++ {
		arm.SetPriority(i, 128)
	}
	// SCB_VTOR = (uint32_t)_VectorsRam;	// use vector table in RAM

	// hardware always starts in FEI mode
	//  C1[CLKS] bits are written to 00
	//  C1[IREFS] bit is written to 1
	//  C6[PLLS] bit is written to 0
	// MCG_SC[FCDIV] defaults to divide by two for internal ref clock
	// I tried changing MSG_SC to divide by 1, it didn't work for me
	// enable capacitors for crystal
	nxp.OSC.CR.Set(nxp.OSC_CR_SC8P | nxp.OSC_CR_SC2P | nxp.OSC_CR_ERCLKEN)
	// enable osc, 8-32 MHz range, low power mode
	nxp.MCG.C2.Set(uint8((2 << nxp.MCG_C2_RANGE_Pos) | nxp.MCG_C2_EREFS))
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

	// now in FBE mode
	//  C1[CLKS] bits are written to 10
	//  C1[IREFS] bit is written to 0
	//  C1[FRDIV] must be written to divide xtal to 31.25-39 kHz
	//  C6[PLLS] bit is written to 0
	//  C2[LP] is written to 0
	// we need faster than the crystal, turn on the PLL (F_CPU > 120000000)
	nxp.SMC.PMCTRL.Set(_SMC_PMCTRL_HSRUN) // enter HSRUN mode
	for nxp.SMC.PMSTAT.Get() != _SMC_PMSTAT_HSRUN {
	} // wait for HSRUN
	nxp.MCG.C5.Set((1 << nxp.MCG_C5_PRDIV_Pos))
	nxp.MCG.C6.Set(nxp.MCG_C6_PLLS | (29 << nxp.MCG_C6_VDIV_Pos))

	// wait for PLL to start using xtal as its input
	for !nxp.MCG.S.HasBits(nxp.MCG_S_PLLST) {
	}
	// wait for PLL to lock
	for !nxp.MCG.S.HasBits(nxp.MCG_S_LOCK0) {
	}
	// now we're in PBE mode

	// now program the clock dividers
	// config divisors: 180 MHz core, 60 MHz bus, 25.7 MHz flash, USB = IRC48M
	nxp.SIM.CLKDIV1.Set((0 << nxp.SIM_CLKDIV1_OUTDIV1_Pos) | (2 << nxp.SIM_CLKDIV1_OUTDIV2_Pos) | (0 << nxp.SIM_CLKDIV1_OUTDIV3_Pos) | (6 << nxp.SIM_CLKDIV1_OUTDIV4_Pos))
	nxp.SIM.CLKDIV2.Set((0 << nxp.SIM_CLKDIV2_USBDIV_Pos))

	// switch to PLL as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set((0 << nxp.MCG_C1_CLKS_Pos) | (4 << nxp.MCG_C1_FRDIV_Pos))
	// wait for PLL clock to be used
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != (3 << nxp.MCG_S_CLKST_Pos) {
	}
	// now we're in PEE mode
	// trace is CPU clock, CLKOUT=OSCERCLK0
	// USB uses IRC48
	nxp.SIM.SOPT2.Set(nxp.SIM_SOPT2_USBSRC | _SIM_SOPT2_IRC48SEL | nxp.SIM_SOPT2_TRACECLKSEL | (6 << nxp.SIM_SOPT2_CLKOUTSEL_Pos))

	// If the RTC oscillator isn't enabled, get it started.  For Teensy 3.6
	// we don't do this early.  See comment above about slow rising power.
	if !nxp.RTC.CR.HasBits(nxp.RTC_CR_OSCE) {
		nxp.RTC.SR.Set(0)
		nxp.RTC.CR.Set(nxp.RTC_CR_SC16P | nxp.RTC_CR_SC4P | nxp.RTC_CR_OSCE)
	}

	// initialize the SysTick counter
	initSysTick()
}

func initInternal() {
	// from: _init_Teensyduino_internal_
	// arm.EnableIRQ(nxp.IRQ_PORTA)
	// arm.EnableIRQ(nxp.IRQ_PORTB)
	// arm.EnableIRQ(nxp.IRQ_PORTC)
	// arm.EnableIRQ(nxp.IRQ_PORTD)
	// arm.EnableIRQ(nxp.IRQ_PORTE)

	nxp.FTM0.CNT.Set(0)
	nxp.FTM0.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.FTM0.C0SC.Set(0x28) // MSnB:MSnA = 10, ELSnB:ELSnA = 10
	nxp.FTM0.C1SC.Set(0x28)
	nxp.FTM0.C2SC.Set(0x28)
	nxp.FTM0.C3SC.Set(0x28)
	nxp.FTM0.C4SC.Set(0x28)
	nxp.FTM0.C5SC.Set(0x28)
	nxp.FTM0.C6SC.Set(0x28)
	nxp.FTM0.C7SC.Set(0x28)

	nxp.FTM3.C0SC.Set(0x28)
	nxp.FTM3.C1SC.Set(0x28)
	nxp.FTM3.C2SC.Set(0x28)
	nxp.FTM3.C3SC.Set(0x28)
	nxp.FTM3.C4SC.Set(0x28)
	nxp.FTM3.C5SC.Set(0x28)
	nxp.FTM3.C6SC.Set(0x28)
	nxp.FTM3.C7SC.Set(0x28)

	nxp.FTM0.SC.Set((1 << nxp.FTM_SC_CLKS_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.FTM_SC_PS_Pos))
	nxp.FTM1.CNT.Set(0)
	nxp.FTM1.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.FTM1.C0SC.Set(0x28)
	nxp.FTM1.C1SC.Set(0x28)
	nxp.FTM1.SC.Set((1 << nxp.FTM_SC_CLKS_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.FTM_SC_PS_Pos))

	// causes a data bus error for unknown reasons
	// nxp.FTM2.CNT.Set(0)
	// nxp.FTM2.MOD.Set(_DEFAULT_FTM_MOD)
	// nxp.FTM2.C0SC.Set(0x28)
	// nxp.FTM2.C1SC.Set(0x28)
	// nxp.FTM2.SC.Set((1 << nxp.FTM_SC_CLKS_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.FTM_SC_PS_Pos))

	nxp.FTM3.CNT.Set(0)
	nxp.FTM3.MOD.Set(_DEFAULT_FTM_MOD)
	nxp.FTM3.C0SC.Set(0x28)
	nxp.FTM3.C1SC.Set(0x28)
	nxp.FTM3.SC.Set((1 << nxp.FTM_SC_CLKS_Pos) | (_DEFAULT_FTM_PRESCALE << nxp.FTM_SC_PS_Pos))

	nxp.SIM.SCGC2.SetBits(nxp.SIM_SCGC2_TPM1)
	nxp.SIM.SOPT2.SetBits((2 << nxp.SIM_SOPT2_TPMSRC_Pos))
	nxp.TPM1.CNT.Set(0)
	nxp.TPM1.MOD.Set(32767)
	nxp.TPM1.C0SC.Set(0x28)
	nxp.TPM1.C1SC.Set(0x28)
	nxp.TPM1.SC.Set((1 << nxp.FTM_SC_CLKS_Pos) | (0 << nxp.FTM_SC_PS_Pos))

	// configure the sleep timer
	initSleepTimer()

	// 	analog_init();
}

func putchar(c byte) {
	machine.PutcharUART(machine.UART0, c)
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
		// machine.PollUSB(&machine.USB0)
		machine.PollUART(machine.UART0)
		machine.PollUART(machine.UART1)
		machine.PollUART(machine.UART2)
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}
