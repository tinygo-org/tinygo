// +build stm32,stm32l5x2

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
	initTIM15()
	machine.UART0.Configure(machine.UARTConfig{})
	initTIM16()
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 1
	PLL_N               = 55
	PLL_P               = 7 // RCC_PLLP_DIV7
	PLL_Q               = 2 // RCC_PLLQ_DIV2
	PLL_R               = 2 // RCC_PLLR_DIV2
)

/*
   clock settings
   +-------------+-----------+
   | LSE         | 32.768khz |
   | SYSCLK      | 110mhz    |
   | HCLK        | 110mhz    |
   | APB1(PCLK1) | 110mhz    |
   | APB2(PCLK2) | 110mhz    |
   +-------------+-----------+
*/
func initCLK() {

	// PWR_CLK_ENABLE
	stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_PWREN)
	_ = stm32.RCC.APB1ENR1.Get()

	// PWR_VOLTAGESCALING_CONFIG
	stm32.PWR.CR1.ReplaceBits(0, stm32.PWR_CR1_VOS_Msk, 0)
	_ = stm32.PWR.CR1.Get()

	// Initialize the High-Speed External Oscillator
	initOsc()

	// Set flash wait states (min 5 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & 0xF) < 5 {
		stm32.FLASH.ACR.ReplaceBits(5, 0xF, 0)
	}

	// Ensure HCLK does not exceed max during transition
	stm32.RCC.CFGR.ReplaceBits(8<<stm32.RCC_CFGR_HPRE_Pos, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set SYSCLK source and wait
	// (3 = RCC_SYSCLKSOURCE_PLLCLK, 2=RCC_CFGR_SWS_Pos)
	stm32.RCC.CFGR.ReplaceBits(3, stm32.RCC_CFGR_SW_Msk, 0)
	for stm32.RCC.CFGR.Get()&(3<<2) != (3 << 2) {
	}

	// Set HCLK
	// (0 = RCC_SYSCLKSOURCE_PLLCLK)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set flash wait states (max 5 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & 0xF) > 5 {
		stm32.FLASH.ACR.ReplaceBits(5, 0xF, 0)
	}

	// Set APB1 and APB2 clocks (0 = DIV1)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_PPRE2_Msk, 0)
}

func initOsc() {
	// Enable HSI, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Disable Backup domain protection
	if !stm32.PWR.CR1.HasBits(stm32.PWR_CR1_DBP) {
		stm32.PWR.CR1.SetBits(stm32.PWR_CR1_DBP)
		for !stm32.PWR.CR1.HasBits(stm32.PWR_CR1_DBP) {
		}
	}

	// Set LSE Drive to LOW
	stm32.RCC.BDCR.ReplaceBits(0, stm32.RCC_BDCR_LSEDRV_Msk, 0)

	// Enable LSE, wait until ready
	stm32.RCC.BDCR.SetBits(stm32.RCC_BDCR_LSEON)
	for !stm32.RCC.BDCR.HasBits(stm32.RCC_BDCR_LSEON) {
	}

	// Ensure LSESYS disabled
	stm32.RCC.BDCR.ClearBits(stm32.RCC_BDCR_LSESYSEN)
	for stm32.RCC.BDCR.HasBits(stm32.RCC_BDCR_LSESYSEN) {
	}

	// Enable HSI48, wait until ready
	stm32.RCC.CRRCR.SetBits(stm32.RCC_CRRCR_HSI48ON)
	for !stm32.RCC.CRRCR.HasBits(stm32.RCC_CRRCR_HSI48ON) {
	}

	// Disable the PLL, wait until disabled
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.ReplaceBits(
		(1)| // 1 = RCC_PLLSOURCE_MSI
			(PLL_M-1)<<stm32.RCC_PLLCFGR_PLLM_Pos|
			(PLL_N<<stm32.RCC_PLLCFGR_PLLN_Pos)|
			(((PLL_Q>>1)-1)<<stm32.RCC_PLLCFGR_PLLQ_Pos)|
			(((PLL_R>>1)-1)<<stm32.RCC_PLLCFGR_PLLR_Pos)|
			(PLL_P<<stm32.RCC_PLLCFGR_PLLPDIV_Pos),
		stm32.RCC_PLLCFGR_PLLSRC_Msk|stm32.RCC_PLLCFGR_PLLM_Msk|
			stm32.RCC_PLLCFGR_PLLN_Msk|stm32.RCC_PLLCFGR_PLLP_Msk|
			stm32.RCC_PLLCFGR_PLLR_Msk|stm32.RCC_PLLCFGR_PLLPDIV_Msk,
		0)

	// Enable the PLL and PLL System Clock Output, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_PLLREN) // = RCC_PLL_SYSCLK
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

// Enable the TIM15 clock.(sleep count)
func initTIM15() {
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM15EN)

	intr := interrupt.New(stm32.IRQ_TIM15, handleTIM15)
	intr.SetPriority(0xc3)
	intr.Enable()
}

// Enable the TIM16 clock.(tick count)
func initTIM16() {
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM16EN)

	// CK_INT = APB1 = 110mhz
	stm32.TIM16.PSC.Set(110000000/10000 - 1) // 110mhz to 10khz(0.1ms)
	stm32.TIM16.ARR.Set(10 - 1)              // interrupt per 1ms

	// Enable the hardware interrupt.
	stm32.TIM16.DIER.SetBits(stm32.TIM_DIER_UIE)

	// Enable the timer.
	stm32.TIM16.CR1.SetBits(stm32.TIM_CR1_CEN)

	intr := interrupt.New(stm32.IRQ_TIM16, handleTIM16)
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

	// CK_INT = APB1 = 110mhz
	// prescale counter down from 110mhz to 10khz aka 0.1 ms frequency.
	stm32.TIM15.PSC.Set(110000000/10000 - 1)

	// set duty aka duration
	arr := (ticks / 100) - 1 // convert from microseconds to 0.1 ms
	if arr == 0 {
		arr = 1 // avoid blocking
	}
	stm32.TIM15.ARR.Set(arr)

	// Enable the hardware interrupt.
	stm32.TIM15.DIER.SetBits(stm32.TIM_DIER_UIE)

	// Enable the timer.
	stm32.TIM15.CR1.SetBits(stm32.TIM_CR1_CEN)

	// wait till timer wakes up
	for timerWakeup.Get() == 0 {
		arm.Asm("wfi")
	}
}

func handleTIM15(interrupt.Interrupt) {
	if stm32.TIM15.SR.HasBits(stm32.TIM_SR_UIF) {
		// Disable the timer.
		stm32.TIM15.CR1.ClearBits(stm32.TIM_CR1_CEN)

		// clear the update flag
		stm32.TIM15.SR.ClearBits(stm32.TIM_SR_UIF)

		// timer was triggered
		timerWakeup.Set(1)
	}
}

func handleTIM16(interrupt.Interrupt) {
	if stm32.TIM16.SR.HasBits(stm32.TIM_SR_UIF) {
		// clear the update flag
		stm32.TIM16.SR.ClearBits(stm32.TIM_SR_UIF)
		tickCount++
	}
}
