//go:build stm32 && stm32l4x5
// +build stm32,stm32l4x5

package runtime

import (
	"device/stm32"
)

/*
clock settings

	+-------------+-----------+
	| LSE         | 32.768khz |
	| SYSCLK      | 120mhz    |
	| HCLK        | 120mhz    |
	| APB1(PCLK1) | 120mhz    |
	| APB2(PCLK2) | 120mhz    |
	+-------------+-----------+
*/
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 1
	PLL_N               = 60
	PLL_P               = RCC_PLLP_DIV2
	PLL_Q               = RCC_PLLQ_DIV2
	PLL_R               = RCC_PLLR_DIV2

	MSIRANGE = stm32.RCC_CR_MSIRANGE_Range4M
)
