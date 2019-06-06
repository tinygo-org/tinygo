// +build sam,atsamd51

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
)

type timeUnit int64

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	callMain()
	abort()
}

func init() {
	initClocks()
	initRTC()
	initSERCOMClocks()
	initUSBClock()
	initADCClock()

	// connect to USB CDC interface
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
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
	sam.GCLK.GENCTRL3.Set((sam.GCLK_GENCTRL_SRC_OSCULP32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL3) {
	}

	// Set OSCULP32K as source of Generic Clock Generator 0
	sam.GCLK.GENCTRL0.Set((sam.GCLK_GENCTRL_SRC_OSCULP32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL0) {
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

	// set GCLK7 to use DFLL48M as clock source
	sam.GCLK.GENCTRL7.Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		(24 << sam.GCLK_GENCTRL_DIVSEL_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL7) {
	}

	// Set up the PLLs

	// Set PLL0 at 120MHz
	sam.GCLK.PCHCTRL1.Set(sam.GCLK_PCHCTRL_CHEN |
		(sam.GCLK_PCHCTRL_GEN_GCLK7 << sam.GCLK_PCHCTRL_GEN_Pos))

	sam.OSCCTRL.DPLLRATIO0.Set((0x0 << sam.OSCCTRL_DPLLRATIO_LDRFRAC_Pos) |
		(59 << sam.OSCCTRL_DPLLRATIO_LDR_Pos))
	for sam.OSCCTRL.DPLLSYNCBUSY0.HasBits(sam.OSCCTRL_DPLLSYNCBUSY_DPLLRATIO) {
	}

	// MUST USE LBYPASS DUE TO BUG IN REV A OF SAMD51, via Adafruit lib.
	sam.OSCCTRL.DPLLCTRLB0.Set((sam.OSCCTRL_DPLLCTRLB_REFCLK_GCLK << sam.OSCCTRL_DPLLCTRLB_REFCLK_Pos) |
		sam.OSCCTRL_DPLLCTRLB_LBYPASS)

	sam.OSCCTRL.DPLLCTRLA0.Set(sam.OSCCTRL_DPLLCTRLA_ENABLE)
	for !sam.OSCCTRL.DPLLSTATUS0.HasBits(sam.OSCCTRL_DPLLSTATUS_CLKRDY) ||
		!sam.OSCCTRL.DPLLSTATUS0.HasBits(sam.OSCCTRL_DPLLSTATUS_LOCK) {
	}

	// // Set PLL1 to 100MHz
	sam.GCLK.PCHCTRL2.Set(sam.GCLK_PCHCTRL_CHEN |
		(sam.GCLK_PCHCTRL_GEN_GCLK7 << sam.GCLK_PCHCTRL_GEN_Pos))

	sam.OSCCTRL.DPLLRATIO1.Set((0x0 << sam.OSCCTRL_DPLLRATIO_LDRFRAC_Pos) |
		(49 << sam.OSCCTRL_DPLLRATIO_LDR_Pos)) // this means 100 Mhz?
	for sam.OSCCTRL.DPLLSYNCBUSY1.HasBits(sam.OSCCTRL_DPLLSYNCBUSY_DPLLRATIO) {
	}

	// // MUST USE LBYPASS DUE TO BUG IN REV A OF SAMD51
	sam.OSCCTRL.DPLLCTRLB1.Set((sam.OSCCTRL_DPLLCTRLB_REFCLK_GCLK << sam.OSCCTRL_DPLLCTRLB_REFCLK_Pos) |
		sam.OSCCTRL_DPLLCTRLB_LBYPASS)

	sam.OSCCTRL.DPLLCTRLA1.Set(sam.OSCCTRL_DPLLCTRLA_ENABLE)
	// for !sam.OSCCTRL.DPLLSTATUS1.HasBits(sam.OSCCTRL_DPLLSTATUS_CLKRDY) ||
	// 	!sam.OSCCTRL.DPLLSTATUS1.HasBits(sam.OSCCTRL_DPLLSTATUS_LOCK) {
	// }

	// Set up the peripheral clocks
	// Set 48MHZ CLOCK FOR USB
	sam.GCLK.GENCTRL1.Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL1) {
	}

	// // Set 100MHZ CLOCK FOR OTHER PERIPHERALS
	// sam.GCLK.GENCTRL2.Set((sam.GCLK_GENCTRL_SRC_DPLL1 << sam.GCLK_GENCTRL_SRC_Pos) |
	// 	sam.GCLK_GENCTRL_IDC |
	// 	sam.GCLK_GENCTRL_GENEN)
	// for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL2) {
	// }

	// // Set 12MHZ CLOCK FOR DAC
	sam.GCLK.GENCTRL4.Set((sam.GCLK_GENCTRL_SRC_DFLL << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		(4 << sam.GCLK_GENCTRL_DIVSEL_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL4) {
	}

	// // Set up main clock
	sam.GCLK.GENCTRL0.Set((sam.GCLK_GENCTRL_SRC_DPLL0 << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	for sam.GCLK.SYNCBUSY.HasBits(sam.GCLK_SYNCBUSY_GENCTRL0) {
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
	sam.OSC32KCTRL.OSCULP32K.Set(sam.OSC32KCTRL_OSCULP32K_EN32K)
	sam.OSC32KCTRL.RTCCTRL.Set(sam.OSC32KCTRL_RTCCTRL_RTCSEL_ULP32K)

	// set Mode0 to 32-bit counter (mode 0) with prescaler 1 and GCLK2 is 32KHz/1
	sam.RTC_MODE0.CTRLA.Set((sam.RTC_MODE0_CTRLA_MODE_COUNT32 << sam.RTC_MODE0_CTRLA_MODE_Pos) |
		(sam.RTC_MODE0_CTRLA_PRESCALER_DIV1 << sam.RTC_MODE0_CTRLA_PRESCALER_Pos) |
		(sam.RTC_MODE0_CTRLA_COUNTSYNC))

	// re-enable RTC
	sam.RTC_MODE0.CTRLA.SetBits(sam.RTC_MODE0_CTRLA_ENABLE)
	for sam.RTC_MODE0.SYNCBUSY.HasBits(sam.RTC_MODE0_SYNCBUSY_ENABLE) {
	}

	arm.SetPriority(sam.IRQ_RTC, 0xc0)
	arm.EnableIRQ(sam.IRQ_RTC)
}

func waitForSync() {
	// for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	// }
}

// treat all ticks params coming from runtime as being in microseconds
const tickMicros = 1000

var (
	timestamp        timeUnit // ticks since boottime
	timerLastCounter uint64
)

//go:volatile
type isrFlag bool

var timerWakeup isrFlag

const asyncScheduler = false

// sleepTicks should sleep for d number of microseconds.
func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks() // update timestamp
		ticks := uint32(d)
		timerSleep(ticks)
		d -= timeUnit(ticks)
	}
}

// ticks returns number of microseconds since start.
func ticks() timeUnit {
	// request read of count
	// sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	// waitForSync()

	rtcCounter := (uint64(sam.RTC_MODE0.COUNT.Get()) * 305) / 10 // each counter tick == 30.5us
	offset := (rtcCounter - timerLastCounter)                    // change since last measurement
	timerLastCounter = rtcCounter
	timestamp += timeUnit(offset) // TODO: not precise
	return timestamp
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup = false
	if ticks < 30 {
		// have to have at least one clock count
		ticks = 30
	}

	// request read of count
	// sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	// waitForSync()

	// set compare value
	cnt := sam.RTC_MODE0.COUNT.Get()
	sam.RTC_MODE0.COMP0.Set(uint32(cnt) + (ticks * 10 / 305)) // each counter tick == 30.5us
	waitForSync()

	// enable IRQ for CMP0 compare
	sam.RTC_MODE0.INTENSET.SetBits(sam.RTC_MODE0_INTENSET_CMP0)

	for !timerWakeup {
		arm.Asm("wfi")
	}
}

//go:export RTC_IRQHandler
func handleRTC() {
	// disable IRQ for CMP0 compare
	sam.RTC_MODE0.INTFLAG.SetBits(sam.RTC_MODE0_INTENSET_CMP0)

	timerWakeup = true
}

func initUSBClock() {
	// Turn on clock(s) for USB
	//MCLK->APBBMASK.reg |= MCLK_APBBMASK_USB;
	//MCLK->AHBMASK.reg |= MCLK_AHBMASK_USB;
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_USB_)
	sam.MCLK.AHBMASK.SetBits(sam.MCLK_AHBMASK_USB_)

	// Put Generic Clock Generator 1 as source for USB
	//GCLK->PCHCTRL[USB_GCLK_ID].reg = GCLK_PCHCTRL_GEN_GCLK1_Val | (1 << GCLK_PCHCTRL_CHEN_Pos);
	sam.GCLK.PCHCTRL10.Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}

func initADCClock() {
	// Turn on clocks for ADC0/ADC1.
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_ADC0_)
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_ADC1_)

	// Put Generic Clock Generator 1 as source for ADC0 and ADC1.
	sam.GCLK.PCHCTRL40.Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL41.Set((sam.GCLK_PCHCTRL_GEN_GCLK1 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}
