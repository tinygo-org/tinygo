// +build sam,atsamd21

package runtime

import (
	"device/arm"
	"device/sam"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
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

	intr := interrupt.New(sam.IRQ_RTC, func(intr interrupt.Interrupt) {
		// disable IRQ for CMP0 compare
		sam.RTC_MODE0.INTFLAG.Set(sam.RTC_MODE0_INTENSET_CMP0)

		timerWakeup.Set(1)
	})
	intr.SetPriority(0xc0)
	intr.Enable()
}

func waitForSync() {
	for sam.GCLK.STATUS.HasBits(sam.GCLK_STATUS_SYNCBUSY) {
	}
}

// treat all ticks params coming from runtime as being in microseconds
const tickMicros = 1000

var (
	timestamp        timeUnit // ticks since boottime
	timerLastCounter uint64
)

var timerWakeup volatile.Register8

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
	sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	waitForSync()

	rtcCounter := (uint64(sam.RTC_MODE0.COUNT.Get()) * 305) / 10 // each counter tick == 30.5us
	offset := (rtcCounter - timerLastCounter)                    // change since last measurement
	timerLastCounter = rtcCounter
	timestamp += timeUnit(offset) // TODO: not precise
	return timestamp
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup.Set(0)
	if ticks < 214 {
		// due to around 183us delay waiting for the register value to sync, the minimum sleep value
		// for the SAMD21 is 214us.
		// For related info, see:
		// https://community.atmel.com/comment/2507091#comment-2507091
		ticks = 214
	}

	// request read of count
	sam.RTC_MODE0.READREQ.Set(sam.RTC_MODE0_READREQ_RREQ)
	waitForSync()

	// set compare value
	cnt := sam.RTC_MODE0.COUNT.Get()
	sam.RTC_MODE0.COMP0.Set(uint32(cnt) + (ticks * 10 / 305)) // each counter tick == 30.5us
	waitForSync()

	// enable IRQ for CMP0 compare
	sam.RTC_MODE0.INTENSET.SetBits(sam.RTC_MODE0_INTENSET_CMP0)

	for timerWakeup.Get() == 0 {
		arm.Asm("wfi")
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
