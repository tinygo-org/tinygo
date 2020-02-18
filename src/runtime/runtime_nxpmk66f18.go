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

// +build nxp,mk66f18

package runtime

import (
	"device/arm"
	"device/nxp"
	"machine"
	"runtime/volatile"
)

const (
	WDOG_UNLOCK_SEQ1 = 0xC520
	WDOG_UNLOCK_SEQ2 = 0xD928

	DEFAULT_FTM_MOD      = 61440 - 1
	DEFAULT_FTM_PRESCALE = 1
)

var (
	SIM_SOPT2_IRC48SEL = nxp.SIM_SOPT2_PLLFLLSEL(3)
	SMC_PMCTRL_HSRUN   = nxp.SMC_PMCTRL_RUNM(3)
	SMC_PMSTAT_HSRUN   = nxp.SMC_PMSTAT_PMSTAT(0x80)
)

var bootMsg = []byte("\r\n\r\nStartup complete, running main\r\n\r\n")

//go:section .resetHandler
//go:export Reset_Handler
func main() {
	initSystem()
	arm.Asm("CPSIE i")
	initInternal()
	startupLateHook()

	initAll()
	machine.UART1.Configure(machine.UARTConfig{BaudRate: 115200})
	for _, c := range bootMsg {
		for !machine.UART1.S1.HasBits(nxp.UART_S1_TDRE) {
		}
		machine.UART1.D.Set(c)
	}

	callMain()
	abort()
}

// ported ResetHandler from mk20dx128.c from teensy3 core libraries
//go:noinline
func initSystem() {
	nxp.WDOG.UNLOCK.Set(WDOG_UNLOCK_SEQ1)
	nxp.WDOG.UNLOCK.Set(WDOG_UNLOCK_SEQ2)
	arm.Asm("nop")
	arm.Asm("nop")
	startupEarlyHook()

	// enable clocks to always-used peripherals
	nxp.SIM.SCGC3.Set(nxp.SIM_SCGC3_ADC1 | nxp.SIM_SCGC3_FTM2 | nxp.SIM_SCGC3_FTM3)
	nxp.SIM.SCGC5.Set(0x00043F82) // clocks active to all GPIO
	nxp.SIM.SCGC6.Set(nxp.SIM_SCGC6_RTC | nxp.SIM_SCGC6_FTM0 | nxp.SIM_SCGC6_FTM1 | nxp.SIM_SCGC6_ADC0 | nxp.SIM_SCGC6_FTF)
	nxp.SystemControl.CPACR.Set(0x00F00000)
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
	nxp.MCG.C2.Set(uint8(nxp.MCG_C2_RANGE(2) | nxp.MCG_C2_EREFS))
	// switch to crystal as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set(uint8(nxp.MCG_C1_CLKS(2) | nxp.MCG_C1_FRDIV(4)))
	// wait for crystal oscillator to begin
	for !nxp.MCG.S.HasBits(nxp.MCG_S_OSCINIT0) {
	}
	// wait for FLL to use oscillator
	for nxp.MCG.S.HasBits(nxp.MCG_S_IREFST) {
	}
	// wait for MCGOUT to use oscillator
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != nxp.MCG_S_CLKST(2) {
	}

	// now in FBE mode
	//  C1[CLKS] bits are written to 10
	//  C1[IREFS] bit is written to 0
	//  C1[FRDIV] must be written to divide xtal to 31.25-39 kHz
	//  C6[PLLS] bit is written to 0
	//  C2[LP] is written to 0
	// we need faster than the crystal, turn on the PLL (F_CPU > 120000000)
	nxp.SMC.PMCTRL.Set(SMC_PMCTRL_HSRUN) // enter HSRUN mode
	for nxp.SMC.PMSTAT.Get() != SMC_PMSTAT_HSRUN {
	} // wait for HSRUN
	nxp.MCG.C5.Set(nxp.MCG_C5_PRDIV(1))
	nxp.MCG.C6.Set(nxp.MCG_C6_PLLS | nxp.MCG_C6_VDIV(29))

	// wait for PLL to start using xtal as its input
	for !nxp.MCG.S.HasBits(nxp.MCG_S_PLLST) {
	}
	// wait for PLL to lock
	for !nxp.MCG.S.HasBits(nxp.MCG_S_LOCK0) {
	}
	// now we're in PBE mode

	// now program the clock dividers
	// config divisors: 180 MHz core, 60 MHz bus, 25.7 MHz flash, USB = IRC48M
	nxp.SIM.CLKDIV1.Set(nxp.SIM_CLKDIV1_OUTDIV1(0) | nxp.SIM_CLKDIV1_OUTDIV2(2) | nxp.SIM_CLKDIV1_OUTDIV4(6))
	nxp.SIM.CLKDIV2.Set(nxp.SIM_CLKDIV2_USBDIV(0))

	// switch to PLL as clock source, FLL input = 16 MHz / 512
	nxp.MCG.C1.Set(nxp.MCG_C1_CLKS(0) | nxp.MCG_C1_FRDIV(4))
	// wait for PLL clock to be used
	for (nxp.MCG.S.Get() & nxp.MCG_S_CLKST_Msk) != nxp.MCG_S_CLKST(3) {
	}
	// now we're in PEE mode
	// trace is CPU clock, CLKOUT=OSCERCLK0
	// USB uses IRC48
	nxp.SIM.SOPT2.Set(nxp.SIM_SOPT2_USBSRC | SIM_SOPT2_IRC48SEL | nxp.SIM_SOPT2_TRACECLKSEL | nxp.SIM_SOPT2_CLKOUTSEL(6))

	// If the RTC oscillator isn't enabled, get it started.  For Teensy 3.6
	// we don't do this early.  See comment above about slow rising power.
	if !nxp.RTC.CR.HasBits(nxp.RTC_CR_OSCE) {
		nxp.RTC.SR.Set(0)
		nxp.RTC.CR.Set(nxp.RTC_CR_SC16P | nxp.RTC_CR_SC4P | nxp.RTC_CR_OSCE)
	}

	// initialize the SysTick counter
	nxp.SysTick.RVR.Set((machine.CPUFrequency() / 1000) - 1)
	nxp.SysTick.CVR.Set(0)
	nxp.SysTick.CSR.Set(nxp.SysTick_CSR_CLKSOURCE | nxp.SysTick_CSR_TICKINT | nxp.SysTick_CSR_ENABLE)
	nxp.SystemControl.SHPR3.Set(0x20200000) // Systick = priority 32
}

// ported _init_Teensyduino_internal_ from pins_teensy.c from teensy3 core libraries
//go:noinline
func initInternal() {
	arm.EnableIRQ(nxp.IRQ_PORTA)
	arm.EnableIRQ(nxp.IRQ_PORTB)
	arm.EnableIRQ(nxp.IRQ_PORTC)
	arm.EnableIRQ(nxp.IRQ_PORTD)
	arm.EnableIRQ(nxp.IRQ_PORTE)

	nxp.FTM0.CNT.Set(0)
	nxp.FTM0.MOD.Set(DEFAULT_FTM_MOD)
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

	nxp.FTM0.SC.Set(nxp.FTM_SC_CLKS(1) | nxp.FTM_SC_PS(DEFAULT_FTM_PRESCALE))
	nxp.FTM1.CNT.Set(0)
	nxp.FTM1.MOD.Set(DEFAULT_FTM_MOD)
	nxp.FTM1.C0SC.Set(0x28)
	nxp.FTM1.C1SC.Set(0x28)
	nxp.FTM1.SC.Set(nxp.FTM_SC_CLKS(1) | nxp.FTM_SC_PS(DEFAULT_FTM_PRESCALE))

	// nxp.FTM2.CNT.Set(0)
	// nxp.FTM2.MOD.Set(DEFAULT_FTM_MOD)
	// nxp.FTM2.C0SC.Set(0x28)
	// nxp.FTM2.C1SC.Set(0x28)
	// nxp.FTM2.SC.Set(nxp.FTM_SC_CLKS(1) | nxp.FTM_SC_PS(DEFAULT_FTM_PRESCALE))

	nxp.FTM3.CNT.Set(0)
	nxp.FTM3.MOD.Set(DEFAULT_FTM_MOD)
	nxp.FTM3.C0SC.Set(0x28)
	nxp.FTM3.C1SC.Set(0x28)
	nxp.FTM3.SC.Set(nxp.FTM_SC_CLKS(1) | nxp.FTM_SC_PS(DEFAULT_FTM_PRESCALE))

	nxp.SIM.SCGC2.SetBits(nxp.SIM_SCGC2_TPM1)
	nxp.SIM.SOPT2.SetBits(nxp.SIM_SOPT2_TPMSRC(2))
	nxp.TPM1.CNT.Set(0)
	nxp.TPM1.MOD.Set(32767)
	nxp.TPM1.C0SC.Set(0x28)
	nxp.TPM1.C1SC.Set(0x28)
	nxp.TPM1.SC.Set(nxp.FTM_SC_CLKS(1) | nxp.FTM_SC_PS(0))

	// configure the low-power timer
	// nxp.LPTMR0.CSR.Set(nxp.LPTMR0_CSR_TIE)
	// nxp.LPTMR0.PSR.Set(nxp.LPTMR0_PSR_PCS(3) | nxp.LPTMR0_PSR_PRESCALE(1)) // use main (external) clock, divided by 4
	// arm.EnableIRQ(nxp.IRQ_LPTMR0)

	// 	analog_init();

	// #if !defined(TEENSY_INIT_USB_DELAY_BEFORE)
	// 	#if TEENSYDUINO >= 142
	// 		#define TEENSY_INIT_USB_DELAY_BEFORE 25
	// 	#else
	// 		#define TEENSY_INIT_USB_DELAY_BEFORE 50
	// 	#endif
	// #endif

	// #if !defined(TEENSY_INIT_USB_DELAY_AFTER)
	// 	#if TEENSYDUINO >= 142
	// 		#define TEENSY_INIT_USB_DELAY_AFTER 275
	// 	#else
	// 		#define TEENSY_INIT_USB_DELAY_AFTER 350
	// 	#endif
	// #endif

	// 	// for background about this startup delay, please see these conversations
	// 	// https://forum.pjrc.com/threads/36606-startup-time-(400ms)?p=113980&viewfull=1#post113980
	// 	// https://forum.pjrc.com/threads/31290-Teensey-3-2-Teensey-Loader-1-24-Issues?p=87273&viewfull=1#post87273

	// 	delay(TEENSY_INIT_USB_DELAY_BEFORE);
	// 	usb_init();
	// 	delay(TEENSY_INIT_USB_DELAY_AFTER);
}

func startupEarlyHook() {
	// TODO allow override
	// > programs using the watchdog timer or needing to initialize hardware as
	// > early as possible can implement startup_early_hook()

	nxp.WDOG.STCTRLH.Set(nxp.WDOG_STCTRLH_ALLOWUPDATE)
}

func startupLateHook() {
	// TODO allow override
}

func putchar(c byte) {
	machine.UART1.WriteByte(c)
}

// ???
const asyncScheduler = false

// microseconds per tick
const tickMicros = 1000

// number of ticks since boot
var tickMilliCount volatile.Register32

//go:export SysTick_Handler
func tickHandler() {
	tickMilliCount.Set(tickMilliCount.Get() + 1)
}

// ticks are in microseconds
func ticks() timeUnit {
	m := arm.DisableInterrupts()
	current := nxp.SysTick.CVR.Get()
	count := tickMilliCount.Get()
	istatus := nxp.SystemControl.ICSR.Get()
	arm.EnableInterrupts(m)

	if istatus&nxp.SystemControl_ICSR_PENDSTSET != 0 && current > 50 {
		count++
	}

	current = ((machine.CPUFrequency() / tickMicros) - 1) - current
	return timeUnit(count*tickMicros + current/(machine.CPUFrequency()/1000000))
}

// sleepTicks spins for a number of microseconds
func sleepTicks(d timeUnit) {
	// TODO actually sleep

	if d <= 0 {
		return
	}

	start := ticks()
	ms := d / 1000

	for {
		for ticks()-start >= 1000 {
			ms--
			if ms <= 0 {
				return
			}
			start += 1000
		}
		arm.Asm("wfi")
	}
}

func Sleep(d int64) {
	sleepTicks(timeUnit(d))
}

// func abort() {
// 	for {
// 		// keep polling some communication while in fault
// 		// mode, so we don't completely die.
// 		if nxp.SIM.SCGC4.HasBits(nxp.SIM_SCGC4_USBOTG) usb_isr();
// 		if nxp.SIM.SCGC4.HasBits(nxp.SIM_SCGC4_UART0) uart0_status_isr();
// 		if nxp.SIM.SCGC4.HasBits(nxp.SIM_SCGC4_UART1) uart1_status_isr();
// 		if nxp.SIM.SCGC4.HasBits(nxp.SIM_SCGC4_UART2) uart2_status_isr();
// 	}
// }
