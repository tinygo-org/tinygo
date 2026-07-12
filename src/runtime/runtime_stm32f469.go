//go:build stm32f4 && stm32f469

package runtime

import (
	"device/stm32"
	"machine"
)

/*
clock settings

	+-------------+--------------------------------+
	| HSE         | selectable (xtal_8/12/16_mhz)   |
	| SYSCLK      | 180mhz                          |
	| HCLK        | 180mhz                          |
	| APB2(PCLK2) | 90mhz                           |
	| APB1(PCLK1) | 45mhz                           |
	+-------------+--------------------------------+
*/
const HSE_STARTUP_TIMEOUT = 0x0500

// pllCFGR builds the RCC_PLLCFGR value for the current xtal - see RM0386
// Reference Manual pg. 148.
func pllCFGR() uint32 {
	pll := machine.PLLParams180MHz()
	return pll.M | (pll.N << stm32.RCC_PLLCFGR_PLLN_Pos) | (((pll.P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | (pll.Q << stm32.RCC_PLLCFGR_PLLQ_Pos) | (pll.R << stm32.RCC_PLLCFGR_PLLR_Pos)
}
