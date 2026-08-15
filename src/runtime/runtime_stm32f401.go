//go:build stm32f401

package runtime

import (
	"device/stm32"
	"machine"
)

const (
	// +---------------------------------------------+
	// |    Clock Settings                           |
	// +-------------+-------------------------------+
	// | HSE         | selectable (xtal_8/16_mhz)    |
	// | SYSCLK      | 84mhz                         |
	// | HCLK        | 84mhz                         |
	// | APB1(PCLK1) | 42mhz                         |
	// | APB2(PCLK2) | 84mhz                         |
	// +-------------+-------------------------------+
	HCLK_FREQ_HZ  = 84000000
	PCLK1_FREQ_HZ = HCLK_FREQ_HZ / 2
	PCLK2_FREQ_HZ = HCLK_FREQ_HZ / 1
)

const (
	// VOS Scale 1 (bits [15:14] = 0b11): max HCLK = 84 MHz
	PWR_SCALE1 = 3 << stm32.PWR_CR_VOS_Pos

	PLL_SRC_HSE = 1 << stm32.RCC_PLLCFGR_PLLSRC_Pos
	PLL_SRC_HSI = 0

	SYSCLK_SRC_PLL  = stm32.RCC_CFGR_SW_PLL << stm32.RCC_CFGR_SW_Pos
	SYSCLK_STAT_PLL = stm32.RCC_CFGR_SWS_PLL << stm32.RCC_CFGR_SWS_Pos

	RCC_DIV_PCLK1 = stm32.RCC_CFGR_PPRE1_Div2 << stm32.RCC_CFGR_PPRE1_Pos // HCLK / 2
	RCC_DIV_PCLK2 = stm32.RCC_CFGR_PPRE2_Div1 << stm32.RCC_CFGR_PPRE2_Pos // HCLK / 1
	RCC_DIV_HCLK  = stm32.RCC_CFGR_HPRE_Div1 << stm32.RCC_CFGR_HPRE_Pos   // SYSCLK / 1
)

const (
	// +-----------------------------------+
	// |    Voltage range = 2.7V - 3.6V    |
	// +----------------+------------------+
	// |   Wait states  |    System Bus    |
	// |  (WS, LATENCY) |    HCLK (MHz)    |
	// +----------------+------------------+
	// | 0 WS, 1 cycle  |   0 < HCLK ≤ 30  |
	// | 1 WS, 2 cycles |  30 < HCLK ≤ 64  |
	// | 2 WS, 3 cycles |  64 < HCLK ≤ 84  |
	// +----------------+------------------+
	FLASH_LATENCY = 2 << stm32.FLASH_ACR_LATENCY_Pos // 2 WS (3 CPU cycles)

	// instruction cache, data cache, and prefetch
	FLASH_OPTIONS = stm32.FLASH_ACR_ICEN | stm32.FLASH_ACR_DCEN | stm32.FLASH_ACR_PRFTEN
)

func init() {
	initOSC()
	initCLK()

	machine.InitSerial()

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

	// disable PLL before configuring it
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// set HSE as PLL source and configure dividers for 84MHz SYSCLK
	pll := machine.PLLParams84MHz()
	stm32.RCC.PLLCFGR.Set(PLL_SRC_HSE |
		pll.M<<stm32.RCC_PLLCFGR_PLLM_Pos |
		pll.N<<stm32.RCC_PLLCFGR_PLLN_Pos |
		((pll.P>>1)-1)<<stm32.RCC_PLLCFGR_PLLP_Pos |
		pll.Q<<stm32.RCC_PLLCFGR_PLLQ_Pos)

	// enable PLL and wait for it to lock
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}

func initCLK() {
	// configure instruction/data caching, prefetch, and flash access wait states
	stm32.FLASH.ACR.Set(FLASH_OPTIONS | FLASH_LATENCY)
	for !stm32.FLASH.ACR.HasBits(FLASH_LATENCY) {
	}

	// set CPU clock source to PLL
	stm32.RCC.CFGR.SetBits(SYSCLK_SRC_PLL)

	// update PCLK1/2 and HCLK divisors
	stm32.RCC.CFGR.SetBits(RCC_DIV_PCLK1 | RCC_DIV_PCLK2 | RCC_DIV_HCLK)

	// verify system clock source is ready
	for !stm32.RCC.CFGR.HasBits(SYSCLK_STAT_PLL) {
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
