//go:build stm32f4 && stm32f469
// +build stm32f4,stm32f469

package runtime

import "tinygo.org/x/device/stm32"

/*
clock settings

	+-------------+--------+
	| HSE         | 8mhz   |
	| SYSCLK      | 180mhz |
	| HCLK        | 180mhz |
	| APB2(PCLK2) | 90mhz  |
	| APB1(PCLK1) | 45mhz  |
	+-------------+--------+
*/
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	// PLL Options - See RM0386 Reference Manual pg. 148
	PLL_M    = 8 // PLL_VCO = (HSE_VALUE or HSI_VALUE / PLL_M) * PLL_N
	PLL_N    = 360
	PLL_P    = 2 // SYSCLK = PLL_VCO / PLL_P
	PLL_Q    = 7 // USB OTS FS, SDIO and RNG Clock = PLL_VCO / PLL_Q
	PLL_R    = 6 // DSI
	PLL_CFGR = PLL_M | (PLL_N << stm32.RCC_PLLCFGR_PLLN_Pos) | (((PLL_P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | (PLL_Q << stm32.RCC_PLLCFGR_PLLQ_Pos) | (PLL_R << stm32.RCC_PLLCFGR_PLLR_Pos)
)
