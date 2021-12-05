//go:build stm32wle5
// +build stm32wle5

package runtime

import (
	"device/stm32"
	"machine"
)

type arrtype = uint32

const (
	/* PLL Options RMN0461.p247 */
	PLL_M = 2
	PLL_N = 6
	PLL_R = 2
	PLL_P = 2
	PLL_Q = 2
)

func init() {

	// Configure 48Mhz clock
	initCLK()

	// UART init
	machine.Serial.Configure(machine.UARTConfig{})

	// Timers init
	initTickTimer(&machine.TIM1)

}

func initCLK() {

	// Enable HSE32 VDDTCXO output on package pin PB0-VDDTCXO
	stm32.RCC.CR.ReplaceBits(stm32.RCC_CR_HSEBYPPWR_VDDTCXO, stm32.RCC_CR_HSEBYPPWR_Msk, 0)

	// SYSCLOCK from HSE32 clock division factor (SYSCLOCK=HSE32)
	stm32.RCC.CR.ReplaceBits(stm32.RCC_CR_HSEPRE_Div1, stm32.RCC_CR_HSEPRE_Msk, 0)

	// enable external Clock HSE32 TXCO (RM0461p226)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEBYPPWR)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Disable PLL
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.Set((PLL_M-1)<<stm32.RCC_PLLCFGR_PLLM_Pos |
		((PLL_N) << stm32.RCC_PLLCFGR_PLLN_Pos) |
		((PLL_Q - 1) << stm32.RCC_PLLCFGR_PLLQ_Pos) |
		((PLL_R - 1) << stm32.RCC_PLLCFGR_PLLR_Pos) |
		((PLL_P - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		stm32.RCC_PLLCFGR_PLLSRC_HSE32 | stm32.RCC_PLLCFGR_PLLPEN | stm32.RCC_PLLCFGR_PLLQEN)

	// Enable PLL
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Enable PLL System Clock output.
	stm32.RCC.PLLCFGR.SetBits(stm32.RCC_PLLCFGR_PLLREN)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Set Flash Latency of 2
	stm32.FLASH.ACR.ReplaceBits(stm32.Flash_ACR_LATENCY_WS2, stm32.Flash_ACR_LATENCY_Msk, 0)
	for (stm32.FLASH.ACR.Get() & stm32.Flash_ACR_LATENCY_Msk) != stm32.Flash_ACR_LATENCY_WS2 {
	}

	// HCLK1 clock division factor
	stm32.RCC.CFGR.ReplaceBits(stm32.RCC_CFGR_HPRE_Div1, stm32.RCC_CFGR_HPRE_Msk, 0)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_HPREF) {
	}

	// HCLK3 clock division factor
	stm32.RCC.EXTCFGR.ReplaceBits(stm32.RCC_EXTCFGR_SHDHPRE_Div1, stm32.RCC_EXTCFGR_SHDHPRE_Msk, 0)
	for !stm32.RCC.EXTCFGR.HasBits(stm32.RCC_EXTCFGR_SHDHPREF) {
	}

	// PCLK1 clock division factor
	stm32.RCC.CFGR.ReplaceBits(stm32.RCC_CFGR_PPRE1_Div1, stm32.RCC_CFGR_PPRE1_Msk, 0)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_PPRE1F) {
	}

	// PCLK2 clock division factor
	stm32.RCC.CFGR.ReplaceBits(stm32.RCC_CFGR_PPRE2_Div1, stm32.RCC_CFGR_PPRE2_Msk, 0)
	for !stm32.RCC.CFGR.HasBits(stm32.RCC_CFGR_PPRE2F) {
	}

	// Set clock source to PLL
	stm32.RCC.CFGR.ReplaceBits(stm32.RCC_CFGR_SW_PLLR, stm32.RCC_CFGR_SW_Msk, 0)
	for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != (stm32.RCC_CFGR_SWS_PLLR << stm32.RCC_CFGR_SWS_Pos) {
	}

}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}
