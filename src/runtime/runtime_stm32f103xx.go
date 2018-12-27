// +build stm32,stm32f103xx

package runtime

import (
	"device/arm"
	"device/stm32"
	"machine"
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

// initCLK sets clock to 72MHz using HSE 8MHz crystal w/ PLL X 9 (8MHz x 9 = 72MHz).
func initCLK() {
	stm32.FLASH.ACR |= stm32.FLASH_ACR_LATENCY_2 // Two wait states, per datasheet
	stm32.RCC.CFGR |= stm32.RCC_CFGR_PPRE1_DIV_2 // prescale AHB1 = HCLK/2
	stm32.RCC.CR |= stm32.RCC_CR_HSEON           // enable HSE clock

	// wait for the HSEREADY flag
	for (stm32.RCC.CR & stm32.RCC_CR_HSERDY) == 0 {
	}

	stm32.RCC.CFGR |= stm32.RCC_CFGR_PLLSRC   // set PLL source to HSE
	stm32.RCC.CFGR |= stm32.RCC_CFGR_PLLMUL_9 // multiply by 9
	stm32.RCC.CR |= stm32.RCC_CR_PLLON        // enable the PLL

	// wait for the PLLRDY flag
	for (stm32.RCC.CR & stm32.RCC_CR_PLLRDY) == 0 {
	}

	stm32.RCC.CFGR |= stm32.RCC_CFGR_SW_PLL // set clock source to pll

	// wait for PLL to be CLK
	for (stm32.RCC.CFGR & stm32.RCC_CFGR_SWS_PLL) == 0 {
	}
}

const tickMicros = 1024 * 32 // how many clock "ticks" in one microsecond?
//const tickMicros = 1 // 1024 * 32

var (
	timestamp        timeUnit // nanoseconds since boottime
	timerLastCounter uint32   // 24 bits ticks
)

//go:volatile
type isrFlag bool

var timerWakeup isrFlag

func initRTC() {
	// Enable the PWR and BKP.
	stm32.RCC.APB1ENR |= stm32.RCC_APB1ENR_PWREN | stm32.RCC_APB1ENR_BKPEN

	// access to backup register
	stm32.PWR.CR |= stm32.PWR_CR_DBP

	// Enable LSE
	stm32.RCC.BDCR |= stm32.RCC_BDCR_LSEON

	// wait until LSE is ready
	for stm32.RCC.BDCR&stm32.RCC_BDCR_LSERDY == 0 {
	}

	// Select LSE
	stm32.RCC.BDCR |= stm32.RCC_RTCCLKSource_LSE

	// set prescaler to "max" per datasheet
	stm32.RTC.PRLH = 0x000f
	stm32.RTC.PRLL = 0x7FFF

	// set count to zero
	stm32.RTC.CNTH = 0x0
	stm32.RTC.CNTL = 0x0

	// Enable RTC
	stm32.RCC.BDCR |= stm32.RCC_BDCR_RTCEN

	// Clear RSF
	stm32.RTC.CRL &^= stm32.RTC_CRL_RSF

	// Wait till flag is set
	for stm32.RTC.CRL&stm32.RTC_CRL_RSF == 0 {
	}

	// arm.SetPriority(stm32.IRQ_RTC, 0xc0) // low priority
	// arm.EnableIRQ(stm32.IRQ_RTC)
}

func initTIM() {
	// Enable the TIM3 clock.
	stm32.RCC.APB1ENR |= stm32.RCC_APB1ENR_TIM3EN

	//arm.SetPriority(stm32.IRQ_TIM3, 0xc3)
	arm.EnableIRQ(stm32.IRQ_TIM3)
}

// sleepTicks should sleep for specific number of nanoseconds.
func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks()                       // update timestamp
		ticks := uint32(d) & 0x7fffff // 23 bits (to be on the safe side)
		timerSleep(ticks)             // TODO: not accurate (must be d / 30.5175...)
		d -= timeUnit(ticks)
	}
}

// number of ticks (ns) since start.
func ticks() timeUnit {
	// convert RTC counter from seconds to ns
	timerCounter := uint32(stm32.RTC.CNTH<<16|stm32.RTC.CNTL) * 1000 * 1000 * 1000

	// TODO: add the fractional part of current time using DIV registers

	offset := (timerCounter - timerLastCounter) // change since last measurement
	timerLastCounter = timerCounter
	timestamp += timeUnit(offset)
	return timestamp
}

// ticks are in nsec
func timerSleep(ticks uint32) {
	timerWakeup = false

	// STM32 timer update event period is calculated as follows:
	//
	// 			Update_event = TIM_CLK/((PSC + 1)*(ARR + 1)*(RCR + 1))
	//
	// Where:
	//
	//			TIM_CLK = timer clock input
	// 			PSC = 16-bit prescaler register
	// 			ARR = 16/32-bit Autoreload register
	// 			RCR = 16-bit repetition counter
	//
	// Example:
	//
	//			TIM_CLK = 72 MHz
	// 			Prescaler = 1
	// 			Auto reload = 65535
	// 			No repetition counter RCR = 0
	// 			Update_event = 72*(10^6)/((1 + 1)*(65535 + 1)*(1))
	// 			Update_event = 549.3 Hz
	//
	// Set the timer prescaler/autoreload timing registers.

	// prescale counter down from 72mhz to 10khz aka 0.1 ms frequency.
	stm32.TIM3.PSC = 7199 //machine.CPU_FREQUENCY / 1000 ?

	// set duty aka duration
	//stm32.TIM3.ARR = 4999 // sets to 500 ms (5000 x 0.1ms - 1)
	stm32.TIM3.ARR = stm32.RegValue(ticks/1000*1000*1000*10) - 1 // convert from ns to 0.1ms

	// Enable the hardware interrupt.
	stm32.TIM3.DIER |= stm32.TIM_DIER_UIE

	// Send an update event to reset the timer and apply settings.
	stm32.TIM3.EGR |= stm32.TIM_EGR_UG

	// Enable the timer.
	stm32.TIM3.CR1 |= stm32.TIM_CR1_CEN

	// wait till timer wakes up
	for !timerWakeup {
		arm.Asm("wfi")
	}
}

//go:export TIM3_IRQHandler
func handleTIM3() {
	if (stm32.TIM3.SR & stm32.TIM_SR_UIF) > 0 {
		// Disable the timer.
		stm32.TIM3.CR1 &^= stm32.TIM_CR1_CEN

		// clear the update flag
		stm32.TIM3.SR &^= stm32.TIM_SR_UIF

		// timer was triggered
		timerWakeup = true
	}
}
