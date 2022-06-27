//go:build sam && atsamd21
// +build sam,atsamd21

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
	"machine/usb/cdc"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

type timeUnit int64

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}

func init() {
	initClocks()
	initRTC()
	initSERCOMClocks()
	initUSBClock()
	initADCClock()

	// connect to USB CDC interface
	cdc.EnableUSBCDC()
	machine.USB.Configure(machine.UARTConfig{})
	machine.InitSerial()
	machine.Serial.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

func initClocks() {
	// Set 1 Flash Wait State for 48MHz, required for 3.3V operation according to SAMD21 Datasheet
	sam.NVMCTRL.CTRLB.SetBits(sam.NVMCTRL_CTRLB_RWS_HALF << sam.NVMCTRL_CTRLB_RWS_Pos)

	// Turn on the digital interface clock
	sam.PM.APBAMASK.SetBits(sam.PM_APBAMASK_GCLK_)
	// turn off RTC
	sam.PM.APBAMASK.ClearBits(sam.PM_APBAMASK_RTC_)

	// Enable OSC32K clock (Internal 32.768Hz oscillator).
	// This requires registers that are not included in the SVD file.
	// This is from samd21g18a.h and nvmctrl.h:
	//
	// #define NVMCTRL_OTP4 0x00806020
	//
	// #define SYSCTRL_FUSES_OSC32K_CAL_ADDR (NVMCTRL_OTP4 + 4)
	// #define SYSCTRL_FUSES_OSC32K_CAL_Pos 6 /** (NVMCTRL_OTP4) OSC32K Calibration */
	// #define SYSCTRL_FUSES_OSC32K_CAL_Msk (0x7Fu << SYSCTRL_FUSES_OSC32K_CAL_Pos)
	// #define SYSCTRL_FUSES_OSC32K_CAL(value) ((SYSCTRL_FUSES_OSC32K_CAL_Msk & ((value) << SYSCTRL_FUSES_OSC32K_CAL_Pos)))
	// u32_t fuse = *(u32_t *)FUSES_OSC32K_CAL_ADDR;
	// u32_t calib = (fuse & FUSES_OSC32K_CAL_Msk) >> FUSES_OSC32K_CAL_Pos;
	fuse := *(*uint32)(unsafe.Pointer(uintptr(0x00806020) + 4))
	calib := (fuse & uint32(0x7f<<6)) >> 6

	// SYSCTRL_OSC32K_CALIB(calib) |
	//  SYSCTRL_OSC32K_STARTUP(0x6u) |
	//  SYSCTRL_OSC32K_EN32K | SYSCTRL_OSC32K_ENABLE;
	sam.SYSCTRL.OSC32K.Set((calib << sam.SYSCTRL_OSC32K_CALIB_Pos) |
		(0x6 << sam.SYSCTRL_OSC32K_STARTUP_Pos) |
		sam.SYSCTRL_OSC32K_EN32K |
		sam.SYSCTRL_OSC32K_EN1K |
		sam.SYSCTRL_OSC32K_ENABLE)
	// Wait for oscillator stabilization
	for !sam.SYSCTRL.PCLKSR.HasBits(sam.SYSCTRL_PCLKSR_OSC32KRDY) {
	}

	// Software reset the module to ensure it is re-initialized correctly
	sam.GCLK.CTRL.Set(sam.GCLK_CTRL_SWRST)
	// Wait for reset to complete
	for sam.GCLK.CTRL.HasBits(sam.GCLK_CTRL_SWRST) && sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	}

	// Put OSC32K as source of Generic Clock Generator 1
	sam.GCLK.GENDIV.Set((1 << sam.GCLK_GENDIV_ID_Pos) |
		(0 << sam.GCLK_GENDIV_DIV_Pos))
	waitForSync()

	// GCLK_GENCTRL_ID(1) | GCLK_GENCTRL_SRC_OSC32K | GCLK_GENCTRL_GENEN;
	sam.GCLK.GENCTRL.Set((1 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_OSC32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	waitForSync()

	// Use Generic Clock Generator 1 as source for Generic Clock Multiplexer 0 (DFLL48M reference)
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_DFLL48 << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK1 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// Remove the OnDemand mode, Bug http://avr32.icgroup.norway.atmel.com/bugzilla/show_bug.cgi?id=9905
	sam.SYSCTRL.DFLLCTRL.Set(sam.SYSCTRL_DFLLCTRL_ENABLE)
	// Wait for ready
	for !sam.SYSCTRL.PCLKSR.HasBits(sam.SYSCTRL_PCLKSR_DFLLRDY) {
	}

	// Handle DFLL calibration based on info learned from Arduino SAMD implementation,
	// using value stored in fuse.
	// #define SYSCTRL_FUSES_DFLL48M_COARSE_CAL_ADDR (NVMCTRL_OTP4 + 4)
	// #define SYSCTRL_FUSES_DFLL48M_COARSE_CAL_Pos 26           /**< \brief (NVMCTRL_OTP4) DFLL48M Coarse Calibration */
	// #define SYSCTRL_FUSES_DFLL48M_COARSE_CAL_Msk (0x3Fu << SYSCTRL_FUSES_DFLL48M_COARSE_CAL_Pos)
	// #define SYSCTRL_FUSES_DFLL48M_COARSE_CAL(value) ((SYSCTRL_FUSES_DFLL48M_COARSE_CAL_Msk & ((value) << SYSCTRL_FUSES_DFLL48M_COARSE_CAL_Pos)))
	coarse := (fuse >> 26) & 0x3F
	if coarse == 0x3f {
		coarse = 0x1f
	}

	sam.SYSCTRL.DFLLVAL.SetBits(coarse << sam.SYSCTRL_DFLLVAL_COARSE_Pos)
	sam.SYSCTRL.DFLLVAL.SetBits(0x1ff << sam.SYSCTRL_DFLLVAL_FINE_Pos)

	// Write full configuration to DFLL control register
	// SYSCTRL_DFLLMUL_CSTEP( 0x1f / 4 ) | // Coarse step is 31, half of the max value
	// SYSCTRL_DFLLMUL_FSTEP( 10 ) |
	// SYSCTRL_DFLLMUL_MUL( (48000) ) ;
	sam.SYSCTRL.DFLLMUL.Set(((31 / 4) << sam.SYSCTRL_DFLLMUL_CSTEP_Pos) |
		(10 << sam.SYSCTRL_DFLLMUL_FSTEP_Pos) |
		(48000 << sam.SYSCTRL_DFLLMUL_MUL_Pos))

	// disable DFLL
	sam.SYSCTRL.DFLLCTRL.Set(0)
	waitForSync()

	sam.SYSCTRL.DFLLCTRL.SetBits(sam.SYSCTRL_DFLLCTRL_MODE |
		sam.SYSCTRL_DFLLCTRL_CCDIS |
		sam.SYSCTRL_DFLLCTRL_USBCRM |
		sam.SYSCTRL_DFLLCTRL_BPLCKC)
	// Wait for ready
	for !sam.SYSCTRL.PCLKSR.HasBits(sam.SYSCTRL_PCLKSR_DFLLRDY) {
	}

	// Re-enable the DFLL
	sam.SYSCTRL.DFLLCTRL.SetBits(sam.SYSCTRL_DFLLCTRL_ENABLE)
	// Wait for ready
	for !sam.SYSCTRL.PCLKSR.HasBits(sam.SYSCTRL_PCLKSR_DFLLRDY) {
	}

	// Switch Generic Clock Generator 0 to DFLL48M. CPU will run at 48MHz.
	sam.GCLK.GENDIV.Set((0 << sam.GCLK_GENDIV_ID_Pos) |
		(0 << sam.GCLK_GENDIV_DIV_Pos))
	waitForSync()

	sam.GCLK.GENCTRL.Set((0 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_DFLL48M << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_IDC |
		sam.GCLK_GENCTRL_GENEN)
	waitForSync()

	// Modify PRESCaler value of OSC8M to have 8MHz
	sam.SYSCTRL.OSC8M.SetBits(sam.SYSCTRL_OSC8M_PRESC_0 << sam.SYSCTRL_OSC8M_PRESC_Pos)
	sam.SYSCTRL.OSC8M.ClearBits(1 << sam.SYSCTRL_OSC8M_ONDEMAND_Pos)
	// Wait for oscillator stabilization
	for !sam.SYSCTRL.PCLKSR.HasBits(sam.SYSCTRL_PCLKSR_OSC8MRDY) {
	}

	// Use OSC8M as source for Generic Clock Generator 3
	sam.GCLK.GENDIV.Set((3 << sam.GCLK_GENDIV_ID_Pos))
	waitForSync()

	sam.GCLK.GENCTRL.Set((3 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_OSC8M << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	waitForSync()

	// Use OSC32K as source for Generic Clock Generator 2
	// OSC32K/1 -> GCLK2 at 32KHz
	sam.GCLK.GENDIV.Set(2 << sam.GCLK_GENDIV_ID_Pos)
	waitForSync()

	sam.GCLK.GENCTRL.Set((2 << sam.GCLK_GENCTRL_ID_Pos) |
		(sam.GCLK_GENCTRL_SRC_OSC32K << sam.GCLK_GENCTRL_SRC_Pos) |
		sam.GCLK_GENCTRL_GENEN)
	waitForSync()

	// Use GCLK2 for RTC
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_RTC << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK2 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()

	// Set the CPU, APBA, B, and C dividers
	sam.PM.CPUSEL.Set(sam.PM_CPUSEL_CPUDIV_DIV1)
	sam.PM.APBASEL.Set(sam.PM_APBASEL_APBADIV_DIV1)
	sam.PM.APBBSEL.Set(sam.PM_APBBSEL_APBBDIV_DIV1)
	sam.PM.APBCSEL.Set(sam.PM_APBCSEL_APBCDIV_DIV1)

	// Disable automatic NVM write operations
	sam.NVMCTRL.CTRLB.SetBits(sam.NVMCTRL_CTRLB_MANW)
}

func initRTC() {
	// turn on digital interface clock
	sam.PM.APBAMASK.SetBits(sam.PM_APBAMASK_RTC_)

	// disable RTC
	sam.RTC_MODE0.CTRL.Set(0)
	waitForSync()

	// reset RTC
	sam.RTC_MODE0.CTRL.SetBits(sam.RTC_MODE0_CTRL_SWRST)
	waitForSync()

	// set Mode0 to 32-bit counter (mode 0) with prescaler 1 and GCLK2 is 32KHz/1
	sam.RTC_MODE0.CTRL.Set((sam.RTC_MODE0_CTRL_MODE_COUNT32 << sam.RTC_MODE0_CTRL_MODE_Pos) |
		(sam.RTC_MODE0_CTRL_PRESCALER_DIV1 << sam.RTC_MODE0_CTRL_PRESCALER_Pos))
	waitForSync()

	// re-enable RTC
	sam.RTC_MODE0.CTRL.SetBits(sam.RTC_MODE0_CTRL_ENABLE)
	waitForSync()

	rtcInterrupt := interrupt.New(sam.IRQ_RTC, func(intr interrupt.Interrupt) {
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
	rtcInterrupt.SetPriority(0xc0)
	rtcInterrupt.Enable()
}

func waitForSync() {
	for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
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
			// Bail out early to handle a non-time interrupt.
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
	// request read of count
	sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	waitForSync()

	return sam.RTC_MODE0.COUNT.Get()
}

// ticks are in microseconds
// Returns true if the timer completed.
// Returns false if another interrupt occured which requires an early return to scheduler.
func timerSleep(ticks uint32) bool {
	timerWakeup.Set(0)
	if ticks < 7 {
		// Due to around 6 clock ticks delay waiting for the register value to
		// sync, the minimum sleep value for the SAMD21 is 214us.
		// For related info, see:
		// https://community.atmel.com/comment/2507091#comment-2507091
		ticks = 7
	}

	// request read of count
	sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	waitForSync()

	// set compare value
	cnt := sam.RTC_MODE0.COUNT.Get()
	sam.RTC_MODE0.COMP0.Set(uint32(cnt) + ticks)
	waitForSync()

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
	// Turn on clock for USB
	sam.PM.APBBMASK.SetBits(sam.PM_APBBMASK_USB_)

	// Put Generic Clock Generator 0 as source for Generic Clock Multiplexer 6 (USB reference)
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_USB << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()
}

func initADCClock() {
	// Turn on clock for ADC
	sam.PM.APBCMASK.SetBits(sam.PM_APBCMASK_ADC_)

	// Put Generic Clock Generator 0 as source for Generic Clock Multiplexer for ADC.
	sam.GCLK.CLKCTRL.Set((sam.GCLK_CLKCTRL_ID_ADC << sam.GCLK_CLKCTRL_ID_Pos) |
		(sam.GCLK_CLKCTRL_GEN_GCLK0 << sam.GCLK_CLKCTRL_GEN_Pos) |
		sam.GCLK_CLKCTRL_CLKEN)
	waitForSync()
}

func waitForEvents() {
	arm.Asm("wfe")
}
