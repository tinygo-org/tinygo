// +build stm32,stm32f7x2

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
	initTIM3()
	machine.UART0.Configure(machine.UARTConfig{})
	initTIM7()
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 4
	PLL_N               = 216
	PLL_P               = 2
	PLL_Q               = 2
)

/*
   clock settings
   +-------------+--------+
   | HSE         | 8mhz   |
   | SYSCLK      | 216mhz |
   | HCLK        | 216mhz |
   | APB1(PCLK1) | 27mhz  |
   | APB2(PCLK2) | 108mhz |
   +-------------+--------+
*/
func initCLK() {
	// PWR_CLK_ENABLE
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	_ = stm32.RCC.APB1ENR.Get()

	// PWR_VOLTAGESCALING_CONFIG
	stm32.PWR.CR1.ReplaceBits(0x3<<stm32.PWR_CR1_VOS_Pos, stm32.PWR_CR1_VOS_Msk, 0)
	_ = stm32.PWR.CR1.Get()

	// Initialize the High-Speed External Oscillator
	initOsc()

	// Set flash wait states (min 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) < 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// HCLK (0x1C00 = DIV_16, 0x0 = RCC_SYSCLK_DIV1) - ensure timers remain
	// within spec as the SYSCLK source changes.
	stm32.RCC.CFGR.ReplaceBits(0x00001C00, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x00001C00<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set SYSCLK source and wait
	// (2 = PLLCLK, 3 = RCC_CFGR_SW mask, 3 << 3 = RCC_CFGR_SWS mask)
	stm32.RCC.CFGR.ReplaceBits(2, 3, 0)
	for stm32.RCC.CFGR.Get()&(3<<2) != (2 << 2) {
	}

	// Set flash wait states (max 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) > 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// Set APB1 and APB2 clocks (0x1800 = DIV8, 0x1000 = DIV2)
	stm32.RCC.CFGR.ReplaceBits(0x1800, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x1000<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
}

func initOsc() {
	// Enable HSE, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Disable the PLL, wait until disabled
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.Set(0x20000000 |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | // 1 = HSE
		PLL_M |
		(PLL_N << stm32.RCC_PLLCFGR_PLLN_Pos) |
		(((PLL_P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(PLL_Q << stm32.RCC_PLLCFGR_PLLQ_Pos))

	// Enable the PLL, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}

var (
	// tick in milliseconds
	tickCount timeUnit
)

var timerWakeup volatile.Register8

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

// Enable the TIM3 clock.(sleep count)
func initTIM3() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)

	intr := interrupt.New(stm32.IRQ_TIM3, handleTIM3)
	intr.SetPriority(0xc3)
	intr.Enable()
}

// Enable the TIM7 clock.(tick count)
func initTIM7() {
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM7EN)

	// CK_INT = APB1 x2 = 54mhz
	stm32.TIM7.PSC.Set(54000000/10000 - 1) // 54mhz to 10khz(0.1ms)
	stm32.TIM7.ARR.Set(10 - 1)             // interrupt per 1ms

	// Enable the hardware interrupt.
	stm32.TIM7.DIER.SetBits(stm32.TIM_DIER_UIE)

	// Enable the timer.
	stm32.TIM7.CR1.SetBits(stm32.TIM_CR1_CEN)

	intr := interrupt.New(stm32.IRQ_TIM7, handleTIM7)
	intr.SetPriority(0xc1)
	intr.Enable()
}

const asyncScheduler = false

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
	timerSleep(uint32(d))
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	// milliseconds to microseconds
	return tickCount * 1000
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup.Set(0)

	// CK_INT = APB1 x2 = 54mhz
	// prescale counter down from 54mhz to 10khz aka 0.1 ms frequency.
	stm32.TIM3.PSC.Set(54000000/10000 - 1)

	// set duty aka duration
	arr := (ticks / 100) - 1 // convert from microseconds to 0.1 ms
	if arr == 0 {
		arr = 1 // avoid blocking
	}
	stm32.TIM3.ARR.Set(arr)

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

func handleTIM7(interrupt.Interrupt) {
	if stm32.TIM7.SR.HasBits(stm32.TIM_SR_UIF) {
		// clear the update flag
		stm32.TIM7.SR.ClearBits(stm32.TIM_SR_UIF)
		tickCount++
	}
}
