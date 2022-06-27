//go:build stm32f405
// +build stm32f405

package runtime

import (
	"device/stm32"
	"machine"
)

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

	PLL_DIV_M = 6 << stm32.RCC_PLLCFGR_PLLM_Pos
	PLL_MLT_N = 168 << stm32.RCC_PLLCFGR_PLLN_Pos
	PLL_DIV_P = ((2 >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos
	PLL_DIV_Q = 7 << stm32.RCC_PLLCFGR_PLLQ_Pos

	SYSCLK_SRC_PLL  = stm32.RCC_CFGR_SW_PLL << stm32.RCC_CFGR_SW_Pos
	SYSCLK_STAT_PLL = stm32.RCC_CFGR_SWS_PLL << stm32.RCC_CFGR_SWS_Pos

	RCC_DIV_PCLK1 = stm32.RCC_CFGR_PPRE1_Div4 << stm32.RCC_CFGR_PPRE1_Pos // HCLK / 4
	RCC_DIV_PCLK2 = stm32.RCC_CFGR_PPRE2_Div2 << stm32.RCC_CFGR_PPRE2_Pos // HCLK / 2
	RCC_DIV_HCLK  = stm32.RCC_CFGR_HPRE_Div1 << stm32.RCC_CFGR_HPRE_Pos   // SYSCLK / 1

	CLK_CCM_RAM = 1 << 20
)

const (
	// +-----------------------------------+
	// |    Voltage range = 2.7V - 3.6V    |
	// +----------------+------------------+
	// |   Wait states  |    System Bus    |
	// |  (WS, LATENCY) |    HCLK (MHz)    |
	// +----------------+------------------+
	// | 0 WS, 1 cycle  |   0 < HCLK ≤ 30  |
	// | 1 WS, 2 cycles |  30 < HCLK ≤ 60  |
	// | 2 WS, 3 cycles |  60 < HCLK ≤ 90  |
	// | 3 WS, 4 cycles |  90 < HCLK ≤ 120 |
	// | 4 WS, 5 cycles | 120 < HCLK ≤ 150 |
	// | 5 WS, 6 cycles | 150 < HCLK ≤ 168 |
	// +----------------+------------------+
	FLASH_LATENCY = 5 << stm32.FLASH_ACR_LATENCY_Pos // 5 WS (6 CPU cycles)

	// instruction cache, data cache, and prefetch
	FLASH_OPTIONS = stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN
)

func init() {
	initOSC() // configure oscillators
	initCLK()

	initCOM()

	initTickTimer(&machine.TIM3)
}

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
	for !stm32.FLASH.ACR.HasBits(FLASH_LATENCY) { // verify new wait states
	}

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

	// update PCKL1/2 and HCLK divisors
	stm32.RCC.CFGR.SetBits(RCC_DIV_PCLK1 | RCC_DIV_PCLK2 | RCC_DIV_HCLK)

	// verify system clock source is ready
	for !stm32.RCC.CFGR.HasBits(SYSCLK_STAT_PLL) {
	}

	// enable the CCM RAM clock
	stm32.RCC.AHB1ENR.SetBits(CLK_CCM_RAM)
}

func initCOM() {
	if machine.NUM_UART_INTERFACES > 0 {
		machine.InitSerial()
		machine.Serial.Configure(machine.UARTConfig{})
	}
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
