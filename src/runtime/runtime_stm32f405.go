// +build stm32,stm32f4,stm32f405

package runtime

import (
	"device/arm"
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
)

func init() {
	initOSC() // configure oscillators
	initCLK() // configure CPU, AHB, and APB bus clocks
	initTIM() // configure timers
}

const (
	// +----------------------+
	// |    Clock Settings    |
	// +-------------+--------+
	// | HSE         | 12mhz  |
	// | SYSCLK      | 168mhz |
	// | HCLK        | 168mhz |
	// | APB1(PCLK1) | 42mhz  |
	// | APB2(PCLK2) | 84mhz  |
	// +-------------+--------+
	HCLK_FREQ_HZ  = 168000000
	PCLK1_FREQ_HZ = HCLK_FREQ_HZ / 4
	PCLK2_FREQ_HZ = HCLK_FREQ_HZ / 2
)

const (
	PWR_SCALE1 = 1 << stm32.PWR_CSR_VOSRDY_Pos // max value of HCLK = 168 MHz
	PWR_SCALE2 = 0                             // max value of HCLK = 144 MHz

	PLL_SRC_HSE = 1 << stm32.RCC_PLLCFGR_PLLSRC_Pos // use HSE for PLL and PLLI2S
	PLL_SRC_HSI = 0                                 // use HSI for PLL and PLLI2S

	PLL_DIV_M = 6 << stm32.RCC_PLLCFGR_PLLM0_Pos
	PLL_MLT_N = 168 << stm32.RCC_PLLCFGR_PLLN0_Pos
	PLL_DIV_P = ((2 >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP0_Pos
	PLL_DIV_Q = 7 << stm32.RCC_PLLCFGR_PLLQ0_Pos

	SYSCLK_SRC_PLL = 2 << stm32.RCC_CFGR_SW0_Pos

	RCC_DIV_PCLK1 = 5 << stm32.RCC_CFGR_PPRE1_Pos // HCLK / 4
	RCC_DIV_PCLK2 = 4 << stm32.RCC_CFGR_PPRE2_Pos // HCLK / 2
	RCC_DIV_HCLK  = 0 << stm32.RCC_CFGR_HPRE_Pos  // SYSCLK / 1

	CLK_CCM_RAM = 1 << 20
)

const (
	// +---------------------+---------------------------------------------------------------------------+
	// |                     |                                HCLK (MHz)                                 |
	// |                     +------------------+------------------+------------------+------------------+
	// |  Wait states (WS)   |   Voltage range  |   Voltage range  |   Voltage range  |   Voltage range  |
	// |     (LATENCY)       |   2.7 V - 3.6 V  |   2.4 V - 2.7 V  |   2.1 V - 2.4 V  |   1.8 V - 2.1 V  |
	// |                     |                  |                  |                  |   Prefetch OFF   |
	// +---------------------+------------------+------------------+------------------+------------------+
	// | 0 WS (1 CPU cycle)  |   0 < HCLK ≤ 30  |   0 < HCLK ≤ 24  |   0 < HCLK ≤ 22  |   0 < HCLK ≤ 20  |
	// | 1 WS (2 CPU cycles) |  30 < HCLK ≤ 60  |  24 < HCLK ≤ 48  |  22 < HCLK ≤ 44  |  20 < HCLK ≤ 40  |
	// | 2 WS (3 CPU cycles) |  60 < HCLK ≤ 90  |  48 < HCLK ≤ 72  |  44 < HCLK ≤ 66  |  40 < HCLK ≤ 60  |
	// | 3 WS (4 CPU cycles) |  90 < HCLK ≤ 120 |  72 < HCLK ≤ 96  |  66 < HCLK ≤ 88  |  60 < HCLK ≤ 80  |
	// | 4 WS (5 CPU cycles) | 120 < HCLK ≤ 150 |  96 < HCLK ≤ 120 |  88 < HCLK ≤ 110 |  80 < HCLK ≤ 100 |
	// | 5 WS (6 CPU cycles) | 150 < HCLK ≤ 168 | 120 < HCLK ≤ 144 | 110 < HCLK ≤ 132 | 100 < HCLK ≤ 120 |
	// | 6 WS (7 CPU cycles) |                  | 144 < HCLK ≤ 168 | 132 < HCLK ≤ 154 | 120 < HCLK ≤ 140 |
	// | 7 WS (8 CPU cycles) |                  |                  | 154 < HCLK ≤ 168 | 140 < HCLK ≤ 160 |
	// +---------------------+------------------+------------------+------------------+------------------+
	FLASH_LATENCY = 5 << stm32.FLASH_ACR_LATENCY_Pos // 5 WS (6 CPU cycles)

	// instruction cache, data cache, and prefetch
	FLASH_OPTIONS = stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN
)

func initOSC() {
	// enable voltage regulator
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	stm32.PWR.CR.SetBits(PWR_SCALE1)

	// enable HSE
	stm32.RCC.CR.Set(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Since the main-PLL configuration parameters cannot be changed once PLL is
	// enabled, it is recommended to configure PLL before enabling it (selection
	// of the HSI or HSE oscillator as PLL clock source, and configuration of
	// division factors M, N, P, and Q).

	// disable PLL and wait for it to reset
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// set HSE as PLL source and configure clock divisors
	stm32.RCC.PLLCFGR.Set(PLL_SRC_HSE | PLL_DIV_M | PLL_MLT_N | PLL_DIV_P | PLL_DIV_Q)

	// enable PLL and wait for it to sync
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}

func initCLK() {
	// After reset, the CPU clock frequency is 16 MHz and 0 wait state (WS) is
	// configured in the FLASH_ACR register.
	//
	// It is highly recommended to use the following software sequences to tune
	// the number of wait states needed to access the Flash memory with the CPU
	// frequency.
	//
	// 1. Program the new number of wait states to the LATENCY bits in the
	//    FLASH_ACR register
	// 2. Check that the new number of wait states is taken into account to access
	//    the Flash memory by reading the FLASH_ACR register
	// 3. Modify the CPU clock source by writing the SW bits in the RCC_CFGR
	//    register
	// 4. If needed, modify the CPU clock prescaler by writing the HPRE bits in
	//    RCC_CFGR
	// 5. Check that the new CPU clock source or/and the new CPU clock prescaler
	//    value is/are taken into account by reading the clock source status (SWS
	//    bits) or/and the AHB prescaler value (HPRE bits), respectively, in the
	//    RCC_CFGR register.

	// configure instruction/data caching, prefetch, and flash access wait states
	stm32.FLASH.ACR.Set(FLASH_OPTIONS | FLASH_LATENCY)

	// After a system reset, the HSI oscillator is selected as the system clock.
	// When a clock source is used directly or through PLL as the system clock, it
	// is not possible to stop it.
	//
	// A switch from one clock source to another occurs only if the target clock
	// source is ready (clock stable after startup delay or PLL locked). If a
	// clock source that is not yet ready is selected, the switch occurs when the
	// clock source is ready. Status bits in the RCC clock control register
	// (RCC_CR) indicate which clock(s) is (are) ready and which clock is
	// currently used as the system clock.

	// set CPU clock source to PLL
	stm32.RCC.CFGR.SetBits(SYSCLK_SRC_PLL)
	for !stm32.RCC.CFGR.HasBits(SYSCLK_SRC_PLL) {
	}

	// update PCKL1/2 and HCLK divisors
	stm32.RCC.CFGR.SetBits(RCC_DIV_PCLK1 | RCC_DIV_PCLK2 | RCC_DIV_HCLK)

	// enable the CCM RAM clock
	stm32.RCC.AHB1ENR.SetBits(CLK_CCM_RAM)
}

func initTIM() {
	// enable sleep counter (TIM3)
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)

	tim3 := interrupt.New(stm32.IRQ_TIM3, handleTIM3)
	tim3.SetPriority(0xC3)
	tim3.Enable()

	// enable tick counter (TIM7)
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM7EN)

	stm32.TIM7.PSC.Set((PCLK1_FREQ_HZ*2)/10000 - 1) // 84mhz to 10khz(0.1ms)
	stm32.TIM7.ARR.Set(10 - 1)                      // interrupt per 1ms

	stm32.TIM7.DIER.SetBits(stm32.TIM_DIER_UIE) // enable interrupt
	stm32.TIM7.CR1.SetBits(stm32.TIM_CR1_CEN)   // enable timer

	tim7 := interrupt.New(stm32.IRQ_TIM7, handleTIM7)
	tim7.SetPriority(0xC1)
	tim7.Enable()
}

var (
	// tick in milliseconds
	tickCount   timeUnit
	timerWakeup volatile.Register8
)

const asyncScheduler = false

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

// sleepTicks should sleep for specific number of microseconds.
func sleepTicks(d timeUnit) {
	timerSleep(uint32(d))
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	return tickCount * 1000 // milliseconds to microseconds
}

// ticks are in microseconds
func timerSleep(ticks uint32) {
	timerWakeup.Set(0)

	stm32.TIM3.PSC.Set((PCLK1_FREQ_HZ*2)/10000 - 1) // 8399
	arr := (ticks / 100) - 1                        // convert from microseconds to 0.1 ms
	if arr == 0 {
		arr = 1 // avoid blocking
	}
	stm32.TIM3.ARR.Set(arr)

	stm32.TIM3.DIER.SetBits(stm32.TIM_DIER_UIE) // enable interrupt
	stm32.TIM3.CR1.SetBits(stm32.TIM_CR1_CEN)   // enable the timer

	// wait for timer
	for timerWakeup.Get() == 0 {
		arm.Asm("wfi")
	}
}

func handleTIM3(interrupt.Interrupt) {
	if stm32.TIM3.SR.HasBits(stm32.TIM_SR_UIF) {
		stm32.TIM3.CR1.ClearBits(stm32.TIM_CR1_CEN) // disable the timer
		stm32.TIM3.SR.ClearBits(stm32.TIM_SR_UIF)   // clear the update flag
		timerWakeup.Set(1)                          // flag timer ISR
	}
}

func handleTIM7(interrupt.Interrupt) {
	if stm32.TIM7.SR.HasBits(stm32.TIM_SR_UIF) {
		stm32.TIM7.SR.ClearBits(stm32.TIM_SR_UIF) // clear the update flag
		tickCount++
	}
}

func putchar(c byte) {}
