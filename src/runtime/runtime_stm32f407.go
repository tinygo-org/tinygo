//go:build stm32f4 && stm32f407
// +build stm32f4,stm32f407

package runtime

import "device/stm32"

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
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	// PLL Options - See RM0090 Reference Manual pg. 95
	PLL_M    = 8 // PLL_VCO = (HSE_VALUE or HSI_VLAUE / PLL_M) * PLL_N
	PLL_N    = 336
	PLL_P    = 2 // SYSCLK = PLL_VCO / PLL_P
	PLL_Q    = 7 // USB OTS FS, SDIO and RNG Clock = PLL_VCO / PLL_Q
	PLL_CFGR = PLL_M | (PLL_N << stm32.RCC_PLLCFGR_PLLN_Pos) | (((PLL_P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | (PLL_Q << stm32.RCC_PLLCFGR_PLLQ_Pos)
)
