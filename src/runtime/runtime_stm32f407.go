// +build stm32,stm32f407

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
	/* PLL Options - See RM0090 Reference Manual pg. 95 */
	PLL_M = 8 /* PLL_VCO = (HSE_VALUE or HSI_VLAUE / PLL_M) * PLL_N */
	PLL_N = 336
	PLL_P = 2 /* SYSCLK = PLL_VCO / PLL_P */
	PLL_Q = 7 /* USB OTS FS, SDIO and RNG Clock = PLL_VCO / PLL_Q */
)

/*
   clock settings
   +-------------+--------+
   | HSE         | 8mhz   |
   | SYSCLK      | 168mhz |
   | HCLK        | 168mhz |
   | APB2(PCLK2) | 84mhz  |
   | APB1(PCLK1) | 42mhz  |
   +-------------+--------+
*/
func initCLK() {

	// Reset clock registers
	// Set HSION
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Reset CFGR
	stm32.RCC.CFGR.Set(0x00000000)
	// Reset HSEON, CSSON and PLLON
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEON | stm32.RCC_CR_CSSON | stm32.RCC_CR_PLLON)
	// Reset PLLCFGR
	stm32.RCC.PLLCFGR.Set(0x24003010)
	// Reset HSEBYP
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSEBYP)
	// Disable all interrupts
	stm32.RCC.CIR.Set(0x00000000)

	// Set up the clock
	var startupCounter uint32 = 0

	// Enable HSE
	stm32.RCC.CR.Set(stm32.RCC_CR_HSEON)

	// Wait till HSE is ready and if timeout is reached exit
	for {
		startupCounter++
		if stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) || (startupCounter == HSE_STARTUP_TIMEOUT) {
			break
		}
	}
	if stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
		// Enable high performance mode, System frequency up to 168MHz
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
		stm32.PWR.CR.SetBits(0x4000) // PWR_CR_VOS
		// HCLK = SYSCLK / 1
		stm32.RCC.CFGR.SetBits(0x0 << stm32.RCC_CFGR_HPRE_Pos)
		// PCLK2 = HCLK / 2
		stm32.RCC.CFGR.SetBits(0x4 << stm32.RCC_CFGR_PPRE2_Pos)
		// PCLK1 = HCLK / 4
		stm32.RCC.CFGR.SetBits(0x5 << stm32.RCC_CFGR_PPRE1_Pos)
		// Configure the main PLL
		// PLL Options - See RM0090 Reference Manual pg. 95
		stm32.RCC.PLLCFGR.Set(PLL_M | (PLL_N << 6) | (((PLL_P >> 1) - 1) << 16) |
			(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | (PLL_Q << 24))
		// Enable main PLL
		stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
		// Wait till the main PLL is ready
		for (stm32.RCC.CR.Get() & stm32.RCC_CR_PLLRDY) == 0 {
		}
		// Configure Flash prefetch, Instruction cache, Data cache and wait state
		stm32.FLASH.ACR.Set(stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | (5 << stm32.FLASH_ACR_LATENCY_Pos))
		// Select the main PLL as system clock source
		stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_SW0 | stm32.RCC_CFGR_SW1)
		stm32.RCC.CFGR.SetBits(0x2 << stm32.RCC_CFGR_SW0_Pos)
		for (stm32.RCC.CFGR.Get() & (0x3 << stm32.RCC_CFGR_SWS0_Pos)) != (0x2 << stm32.RCC_CFGR_SWS0_Pos) {
		}

	} else {
		// If HSE failed to start up, the application will have wrong clock configuration
		for {
		}
	}
	// Enable the CCM RAM clock
	stm32.RCC.AHB1ENR.SetBits(1 << 20)

}

const tickMicros = 1000

var (
	// tick in milliseconds
	tickCount timeUnit
)

var timerWakeup volatile.Register8

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

	// CK_INT = APB1 x2 = 84mhz
	stm32.TIM7.PSC.Set(84000000/10000 - 1) // 84mhz to 10khz(0.1ms)
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

	// CK_INT = APB1 x2 = 84mhz
	// prescale counter down from 84mhz to 10khz aka 0.1 ms frequency.
	stm32.TIM3.PSC.Set(84000000/10000 - 1) // 8399

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
