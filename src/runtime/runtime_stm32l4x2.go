//go:build stm32 && stm32l4x2
// +build stm32,stm32l4x2

package runtime

import (
	"device/stm32"
)

/*
clock settings

	+-------------+-----------+
	| LSE         | 32.768khz |
	| SYSCLK      | 80mhz     |
	| HCLK        | 80mhz     |
	| APB1(PCLK1) | 80mhz     |
	| APB2(PCLK2) | 80mhz     |
	+-------------+-----------+
*/
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 1
	PLL_N               = 40
	PLL_P               = RCC_PLLP_DIV7
	PLL_Q               = RCC_PLLQ_DIV2
	PLL_R               = RCC_PLLR_DIV2

	MSIRANGE = stm32.RCC_CR_MSIRANGE_Range4M
)
