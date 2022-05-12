//go:build (sam && atsamd51) || (sam && atsame5x)
// +build sam,atsamd51 sam,atsame5x

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

//export Reset_Handler
func main() {
	arm.SCB.CPACR.Set(0) // disable FPU if it is enabled
	preinit()
	run()
	exit(0)
}

func init() {
	initClocks()
	initRTC()
	initSERCOMClocks()
	initADCClock()

	//// connect to USB CDC interface
	//machine.Serial.Configure(usb.UARTConfig{})
	//if !machine.USB.Configured() {
	//	machine.USB.Configure(usb.UARTConfig{})
	//}

	machine.InitUSB(0)
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func initClocks() {
	// set flash wait state
	sam.NVMCTRL.CTRLA.SetBits(0 << sam.NVMCTRL_CTRLA_RWS_Pos)

	// software reset
	sam.GCLK.CTRLA.SetBits(sam.GCLK_CTRLA_SWRST)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_SWRST) {
	}

	// Set OSCULP32K as source of Generic Clock Generator 3
	// GCLK->GENCTRL[GENERIC_CLOCK_GENERATOR_XOSC32K].reg = GCLK_GENCTRL_SRC(GCLK_GENCTRL_SRC_OSCULP32K) | GCLK_GENCTRL_GENEN; //generic clock gen 3
	sam.GCLK.GENCTRL[3].Set((sam.GCLK_GENCTRL_SRC_OSCULP32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK3) {
	}

	// Set OSCULP32K as source of Generic Clock Generator 0
	sam.GCLK.GENCTRL[0].Set((sam.GCLK_GENCTRL_SRC_OSCULP32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK0) {
	}

	// Enable DFLL48M clock
	sam.OSCCTRL.DFLLCTRLA.Set(0)
	sam.OSCCTRL.DFLLMUL.Set((0x1 << sam.OSCCTRL_DFLLMUL_CSTEP_Pos) |
		(0x1 << sam.OSCCTRL_DFLLMUL_FSTEP_Pos) |
		(0x0 << sam.OSCCTRL_DFLLMUL_MUL_Pos))
	for sam.OSCCTRL.DFLLSYNC.HasBits(sam.OSCCTRL_DFLLSYNC_DFLLMUL) {
	}

	sam.OSCCTRL.DFLLCTRLB.Set(0)
	for sam.OSCCTRL.DFLLSYNC.HasBits(sam.OSCCTRL_DFLLSYNC_DFLLCTRLB) {
	}

	sam.OSCCTRL.DFLLCTRLA.SetBits(sam.OSCCTRL_DFLLCTRLA_ENABLE)
	for sam.OSCCTRL.DFLLSYNC.HasBits(sam.OSCCTRL_DFLLSYNC_ENABLE) {
	}

	sam.OSCCTRL.DFLLVAL.Set(sam.OSCCTRL.DFLLVAL.Get())
	for sam.OSCCTRL.DFLLSYNC.HasBits(sam.OSCCTRL_DFLLSYNC_DFLLVAL) {
	}

	sam.OSCCTRL.DFLLCTRLB.Set(sam.OSCCTRL_DFLLCTRLB_WAITLOCK |
		sam.OSCCTRL_DFLLCTRLB_CCDIS |
		sam.OSCCTRL_DFLLCTRLB_USBCRM)
	for !sam.OSCCTRL.STATUS.HasBits(sam.OSCCTRL_STATUS_DFLLRDY) {
	}

	// set GCLK7 to run at 2MHz, using DFLL48M as clock source
	// GCLK7 = 48MHz / 24 = 2MHz
	sam.GCLK.GENCTRL[7].Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		(24 << sam.GCLK_GENCTRL_DIV_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK7) {
	}

	// Set up the PLLs

	// Set PLL0 to run at 120MHz, using GCLK7 as clock source
	sam.GCLK.PCHCTRL[1].Set(sam.GCLK_PCHCTRL_CHEN |
		(sam.GCLK_PCHCTRL_GEN_GCLK7 << sam.GCLK_PCHCTRL_GEN_Pos))

	// multiplier = 59 + 1 + (0/32) = 60
	// PLL0 = 2MHz * 60 = 120MHz
	sam.OSCCTRL.DPLL[0].DPLLRATIO.Set((0x0 << sam.OSCCTRL_DPLL_DPLLRATIO_LDRFRAC_Pos) |
		(59 << sam.OSCCTRL_DPLL_DPLLRATIO_LDR_Pos))
	for sam.OSCCTRL.DPLL[0].DPLLSYNCBUSY.HasBits(sam.OSCCTRL_DPLL_DPLLSYNCBUSY_DPLLRATIO) {
	}

	// MUST USE LBYPASS DUE TO BUG IN REV A OF SAMD51, via Adafruit lib.
	sam.OSCCTRL.DPLL[0].DPLLCTRLB.Set((sam.OSCCTRL_DPLL_DPLLCTRLB_REFCLK_GCLK << sam.OSCCTRL_DPLL_DPLLCTRLB_REFCLK_Pos) |
		sam.OSCCTRL_DPLL_DPLLCTRLB_LBYPASS)

	sam.OSCCTRL.DPLL[0].DPLLCTRLA.Set(sam.OSCCTRL_DPLL_DPLLCTRLA_ENABLE)
	for !sam.OSCCTRL.DPLL[0].DPLLSTATUS.HasBits(sam.OSCCTRL_DPLL_DPLLSTATUS_CLKRDY) ||
		!sam.OSCCTRL.DPLL[0].DPLLSTATUS.HasBits(sam.OSCCTRL_DPLL_DPLLSTATUS_LOCK) {
	}

	// Set PLL1 to run at 100MHz, using GCLK7 as clock source
	sam.GCLK.PCHCTRL[2].Set(sam.GCLK_PCHCTRL_CHEN |
		(sam.GCLK_PCHCTRL_GEN_GCLK7 << sam.GCLK_PCHCTRL_GEN_Pos))

	// multiplier = 49 + 1 + (0/32) = 50
	// PLL1 = 2MHz * 50 = 100MHz
	sam.OSCCTRL.DPLL[1].DPLLRATIO.Set((0x0 << sam.OSCCTRL_DPLL_DPLLRATIO_LDRFRAC_Pos) |
		(49 << sam.OSCCTRL_DPLL_DPLLRATIO_LDR_Pos))
	for sam.OSCCTRL.DPLL[1].DPLLSYNCBUSY.HasBits(sam.OSCCTRL_DPLL_DPLLSYNCBUSY_DPLLRATIO) {
	}

	// // MUST USE LBYPASS DUE TO BUG IN REV A OF SAMD51
	sam.OSCCTRL.DPLL[1].DPLLCTRLB.Set((sam.OSCCTRL_DPLL_DPLLCTRLB_REFCLK_GCLK << sam.OSCCTRL_DPLL_DPLLCTRLB_REFCLK_Pos) |
		sam.OSCCTRL_DPLL_DPLLCTRLB_LBYPASS)

	sam.OSCCTRL.DPLL[1].DPLLCTRLA.Set(sam.OSCCTRL_DPLL_DPLLCTRLA_ENABLE)
	// for !sam.OSCCTRL.DPLLSTATUS1.HasBits(sam.OSCCTRL_DPLLSTATUS_CLKRDY) ||
	// 	!sam.OSCCTRL.DPLLSTATUS1.HasBits(sam.OSCCTRL_DPLLSTATUS_LOCK) {
	// }

	// Set up the peripheral clocks
	// Set 48MHZ CLOCK FOR USB
	sam.GCLK.GENCTRL[1].Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK1) {
	}

	// // Set 100MHZ CLOCK FOR OTHER PERIPHERALS
	// sam.GCLK.GENCTRL2.Set((sam.GCLK_GENCTRL_SRC_DPLL1 << sam.GCLK_GENCTRL_SRC_Pos) |
	// 	sam.GCLK_GENCTRL_IDC |
	// 	sam.GCLK_GENCTRL_GENEN)
	// for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL2) {
	// }

	// // Set 12MHZ CLOCK FOR DAC
	sam.GCLK.GENCTRL[4].Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		(4 << sam.GCLK_GENCTRL_DIVSEL_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK4) {
	}

	// // Set up main clock
	sam.GCLK.GENCTRL[0].Set((sam.GCLK_GENCTRL_SRC_DPLL0 << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL_GCLK0) {
	}

	sam.MCLK.CPUDIV.Set(sam.MCLK_CPUDIV_DIV_DIV1)

	// Use the LDO regulator by default
	sam.SUPC.VREG.ClearBits(sam.SUPC_VREG_SEL)

	// Start up the "Debug Watchpoint and Trace" unit, so that we can use
	// it's 32bit cycle counter for timing.
	//CoreDebug->DEMCR |= CoreDebug_DEMCR_TRCENA_Msk;
	//DWT->CTRL |= DWT_CTRL_CYCCNTENA_Msk;
}

func initRTC() {
	// turn on digital interface clock
	sam.MCLK.APBAMASK.SetBits(sam.MCLK_APBAMASK_RTC_)

	// disable RTC
	sam.RTC_MODE0.CTRLA.ClearBits(sam.RTC_MODE0_CTRLA_ENABLE)
	//sam.RTC_MODE0.CTRLA.Set(0)
	for sam.RTC_MODE0.SYNCBUSY.HasBits(sam.RTC_MODE0_SYNCBUSY_ENABLE) {
	}

	// reset RTC
	sam.RTC_MODE0.CTRLA.SetBits(sam.RTC_MODE0_CTRLA_SWRST)
	for sam.RTC_MODE0.SYNCBUSY.HasBits(sam.RTC_MODE0_SYNCBUSY_SWRST) {
	}

	// set to use ulp 32k oscillator
	sam.OSC32KCTRL.OSCULP32K.SetBits(sam.OSC32KCTRL_OSCULP32K_EN32K)
	sam.OSC32KCTRL.RTCCTRL.Set(sam.OSC32KCTRL_RTCCTRL_RTCSEL_ULP32K)

	// set Mode0 to 32-bit counter (mode 0) with prescaler 1 and GCLK2 is 32KHz/1
	sam.RTC_MODE0.CTRLA.Set((sam.RTC_MODE0_CTRLA_MODE_COUNT32 << sam.RTC_MODE0_CTRLA_MODE_Pos) |
		(sam.RTC_MODE0_CTRLA_PRESCALER_DIV1 << sam.RTC_MODE0_CTRLA_PRESCALER_Pos) |
		(sam.RTC_MODE0_CTRLA_COUNTSYNC))

	// re-enable RTC
	sam.RTC_MODE0.CTRLA.SetBits(sam.RTC_MODE0_CTRLA_ENABLE)
	for sam.RTC_MODE0.SYNCBUSY.HasBits(sam.RTC_MODE0_SYNCBUSY_ENABLE) {
	}

	irq := interrupt.New(sam.IRQ_RTC, func(interrupt.Interrupt) {
		flags := sam.RTC_MODE0.INTFLAG.Get()
		if flags&sam.RTC_MODE0_INTENSET_CMP0 != 0 {
			// The timer (for a sleep) has expired.
			timerWakeup.Set(1)
		}
		if flags&sam.RTC_MODE0_INTENSET_OVF != 0 {
			// The 32-bit RTC timer has overflowed.
			rtcOverflows.Set(rtcOverflows.Get() + 1)
		}
		// Mark this interrupt has handled for CMP0 and OVF.
		sam.RTC_MODE0.INTFLAG.Set(sam.RTC_MODE0_INTENSET_CMP0 | sam.RTC_MODE0_INTENSET_OVF)
	})
	sam.RTC_MODE0.INTENSET.Set(sam.RTC_MODE0_INTENSET_OVF)
	irq.SetPriority(0xc0)
	irq.Enable()
}

func waitForSync() {
	for sam.RTC_MODE0.SYNCBUSY.HasBits(sam.RTC_MODE0_SYNCBUSY_COUNT) {
	}
}

var rtcOverflows volatile.Register32 // number of times the RTC wrapped around

var timerWakeup volatile.Register8

// ticksToNanoseconds converts RTC ticks (at 32768Hz) to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ticks * 1e9 / 32768
	return int64(ticks) * 1953125 / 64
}

// nanosecondsToTicks converts nanoseconds to RTC ticks (running at 32768Hz).
func nanosecondsToTicks(ns int64) timeUnit {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ns * 32768 / 1e9
	return timeUnit(ns * 64 / 1953125)
}

// sleepTicks should sleep for d number of microseconds.
func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks := uint32(d)
		if !timerSleep(ticks) {
			return
		}
		d -= timeUnit(ticks)
	}
}

// ticks returns the elapsed time since reset.
func ticks() timeUnit {
	// For some ways of capturing the time atomically, see this thread:
	// https://www.eevblog.com/forum/microcontrollers/correct-timing-by-timer-overflow-count/msg749617/#msg749617
	// Here, instead of re-reading the counter register if an overflow has been
	// detected, we simply try again because that results in smaller code.
	for {
		mask := interrupt.Disable()
		counter := readRTC()
		overflows := rtcOverflows.Get()
		hasOverflow := sam.RTC_MODE0.INTFLAG.Get()&sam.RTC_MODE0_INTENSET_OVF != 0
		interrupt.Restore(mask)

		if hasOverflow {
			// There was an overflow while trying to capture the timer.
			// Try again.
			continue
		}

		// This is a 32-bit timer, so the number of timer overflows forms the
		// upper 32 bits of this timer.
		return timeUnit(overflows)<<32 + timeUnit(counter)
	}
}

func readRTC() uint32 {
	waitForSync()
	return sam.RTC_MODE0.COUNT.Get()
}

// ticks are in microseconds
// Returns true if the timer completed.
// Returns false if another interrupt occured which requires an early return to scheduler.
func timerSleep(ticks uint32) bool {
	timerWakeup.Set(0)
	if ticks < 8 {
		// due to delay waiting for the register value to sync, the minimum sleep value
		// for the SAMD51 is 260us.
		// For related info for SAMD21, see:
		// https://community.atmel.com/comment/2507091#comment-2507091
		ticks = 8
	}

	// request read of count
	waitForSync()

	// set compare value
	cnt := sam.RTC_MODE0.COUNT.Get()

	sam.RTC_MODE0.COMP[0].Set(uint32(cnt) + ticks)

	// enable IRQ for CMP0 compare
	sam.RTC_MODE0.INTENSET.Set(sam.RTC_MODE0_INTENSET_CMP0)

wait:
	waitForEvents()
	if timerWakeup.Get() != 0 {
		return true
	}
	if hasScheduler {
		// The interurpt may have awoken a goroutine, so bail out early.
		// Disable IRQ for CMP0 compare.
		sam.RTC_MODE0.INTENCLR.Set(sam.RTC_MODE0_INTENSET_CMP0)
		return false
	} else {
		// This is running without a scheduler.
		// The application expects this to sleep the whole time.
		goto wait
	}
}

func initUSBClock() {
	// Turn on clock(s) for USB
	//MCLK->APBBMASK.reg |= MCLK_APBBMASK_USB;
	//MCLK->AHBMASK.reg |= MCLK_AHBMASK_USB;
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_USB_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_USB_)

	// Put Generic Clock Generator 1 as source for USB
	//GCLK->PCHCTRL[USB_GCLK_ID].reg = GCLK_PCHCTRL_GEN_GCLK1_Val | (1 << GCLK_PCHCTRL_CHEN_Pos);
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_USB].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}

func initADCClock() {
	// Turn on clocks for ADC0/ADC1.
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_ADC0_)
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_ADC1_)

	// Put Generic Clock Generator 1 as source for ADC0 and ADC1.
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_ADC0].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_ADC1].Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}

func waitForEvents() {
	arm.Asm("wfe")
}
