// +build stm32,stm32l0x2

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

func init() {
	initCLK()
	initRTC()
	initTIM()
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// initCLK sets clock to 32MHz
// SEE: https://github.com/WRansohoff/STM32x0_timer_example/blob/master/src/main.c

func initCLK() {

	// Set the Flash ACR to use 1 wait-state
	// enable the prefetch buffer and pre-read for performance
	stm32.Flash.ACR.SetBits(stm32.Flash_ACR_LATENCY | stm32.Flash_ACR_PRFTEN | stm32.Flash_ACR_PRE_READ)

	// Set presaclers so half system clock (PCLKx = HCLK/2)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE1_DIV_2)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE2_DIV_2)

	// Enable the HSI16 oscillator, since the L0 series boots to the MSI one.
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSI16ON)

	// Wait for HSI16 to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSI16RDYF) {
	}

	// Configure the PLL to use HSI16 with a PLLDIV of 2 and PLLMUL of 4.
	stm32.RCC.CFGR.SetBits(0b01<<stm32.RCC_CFGR_PLLDIV_Pos | 0b0001<<stm32.RCC_CFGR_PLLMUL_Pos)
	stm32.RCC.CFGR.ClearBits(0b10<<stm32.RCC_CFGR_PLLDIV_Pos | 0b1110<<stm32.RCC_CFGR_PLLMUL_Pos)
	stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_PLLSRC)

	// Enable PLL
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)

	// Wait for PLL to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Use PLL As System clock
	stm32.RCC.CFGR.SetBits(0b11)

}

var (
	timestamp        timeUnit // microseconds since boottime
	timerLastCounter uint64
)

var timerWakeup volatile.Register8

func initRTC() {

	// Enable power
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)

	// access to backup register
	stm32.PWR.CR.SetBits(stm32.PWR_CR_DBP)

	// Enable LSE
	stm32.RCC.CSR.SetBits(stm32.RCC_CSR_LSEON)

	// wait until LSE is ready
	for !stm32.RCC.CSR.HasBits(stm32.RCC_CSR_LSERDY) {
	}

	// Select Clock Source LSE
	stm32.RCC.CSR.SetBits(0b01 << stm32.RCC_CSR_RTCSEL_Pos)
	stm32.RCC.CSR.ClearBits(0b10 << stm32.RCC_CSR_RTCSEL_Pos)

	// Enable clock
	stm32.RCC.CSR.SetBits(stm32.RCC_CSR_RTCEN)

	stm32.RTC.WPR.Set(0xCA)                   // Enable Write Access for RTC Registers
	stm32.RTC.WPR.Set(0x53)                   // Enable Write Access for RTC Registers
	stm32.RTC.ISR.SetBits(stm32.RTC_ISR_INIT) // Enable init phase

	// Wait for initialization state
	for !stm32.RTC.ISR.HasBits(stm32.RTC_ISR_INITF) {
	}

	stm32.RTC.PRER.Set(0x003F0270) // set prescaler, 40kHz/64 => 625Hz, 625Hz/625 => 1Hz

	// Set initial date
	//RTC->TR = RTC_TR_PM | 0;

	stm32.RTC.ISR.ClearBits(stm32.RTC_ISR_INIT) // Disable init phase
	stm32.RTC.WPR.Set(0xFE)                     // Disable Write Access for RTC Registers
	stm32.RTC.WPR.Set(0x64)                     // Disable Write Access for RTC Registers
}

// Enable the TIM3 clock.
func initTIM() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)

	intr := interrupt.New(stm32.IRQ_TIM3, handleTIM3)
	intr.SetPriority(0xc3)
	intr.Enable()
}

const asyncScheduler = false

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks()            // update timestamp
		ticks := uint32(d) // current scaling only supports 100 usec to 6553 msec
		timerSleep(ticks)
		d -= timeUnit(ticks)
	}
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {

	// Read twice to force shadow register cache update
	rSubSec := stm32.RTC.SSR.Get() & stm32.RTC_SSR_SS_Msk
	rSubSec = stm32.RTC.SSR.Get() & stm32.RTC_SSR_SS_Msk
	rDate := stm32.RTC.DR.Get()
	rDate = stm32.RTC.DR.Get()
	rDate++
	rTime := stm32.RTC.TR.Get()
	rTime = stm32.RTC.TR.Get()
	prediv := stm32.RTC.PRER.Get() & stm32.RTC_PRER_PREDIV_S_Msk

	var tsec uint64

	// Timestamp in seconds
	tsec = uint64(((rTime & 0x300000) >> 20) * 36000) // Hours Tens
	tsec += uint64(((rTime & 0xf0000) >> 16) * 3600)  // Hours Units
	tsec += uint64(((rTime & 0x7000) >> 12) * 600)    // Minutes Tens
	tsec += uint64(((rTime & 0xf00) >> 8) * 60)       // Minutes Units
	tsec += uint64(((rTime & 0x70) >> 4) * 10)        // Second Tens
	tsec += uint64(rTime & 0xf)                       // Seconds Units

	//Second fraction in milliseconds
	ssec := uint64((1000 * (prediv - rSubSec)) / (prediv + 1))

	timerCounter := uint64(tsec * 1000) // Timestamp in millis
	timerCounter += ssec                // Add sub-seconds
	timerCounter *= 1000                // Convert to micros

	// change since last measurement
	offset := (timerCounter - timerLastCounter)
	timerLastCounter = timerCounter
	timestamp += timeUnit(offset)
	return timestamp
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup.Set(0)

	// prescale counter down from 32mhz to 10khz aka 0.1 ms frequency.
	clk := machine.CPUFrequency() / 2
	stm32.TIM3.PSC.Set(clk/10000 - 1)

	// Set duty aka duration.
	// STM32 dividers use n-1, i.e. n counts from 0 to n-1.
	// As a result, with these prescaler settings,
	// the minimum allowed duration is 200 microseconds.
	if ticks < 200 {
		ticks = 200
	}
	stm32.TIM3.ARR.Set(ticks/100 - 1) // convert from microseconds to 0.1 ms

	// Enable the hardware interrupt.
	stm32.TIM3.DIER.SetBits(stm32.TIM_DIER_UIE)

	// Enable the timer.
	stm32.TIM3.CR1.SetBits(stm32.TIM_CR1_CEN)

	// wait till timer wakes up
	for timerWakeup.Get() == 0 {
		arm.Asm("wfi")
	}
}

func handleTIM3(interrupt.Interrupt) {
	if stm32.TIM3.SR.HasBits(stm32.TIM_SR_UIF) {
		// Disable the timer.
		stm32.TIM3.CR1.ClearBits(stm32.TIM_CR1_CEN)

		// clear the update flag
		stm32.TIM3.SR.ClearBits(stm32.TIM_SR_UIF)

		// timer was triggered
		timerWakeup.Set(1)
	}
}
