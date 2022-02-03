//go:build stm32l0
// +build stm32l0

package runtime

import (
	"device/stm32"
	"machine"
)

const (
	RCC_SYSCLK_DIV1 = 0 // Needs SVD update (should be stm32.RCC_SYSCLK_DIV1)
)

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func initCLK() {
	// Set Power Regulator to enable max performance (1.8V)
	stm32.PWR.CR.ReplaceBits(1<<stm32.PWR_CR_VOS_Pos, stm32.PWR_CR_VOS_Msk, 0)

	// Calibration (default 0x10)
	stm32.RCC.ICSCR.ReplaceBits(0x10<<stm32.RCC_ICSCR_HSI16TRIM_Pos, stm32.RCC_ICSCR_HSI16TRIM_Msk, 0)

	// Enable the HSI16 oscillator, since the L0 series boots to the MSI one.
	stm32.RCC.CR.ReplaceBits(stm32.RCC_CR_HSI16ON, stm32.RCC_CR_HSI16ON_Msk|stm32.RCC_CR_HSI16DIVEN_Msk, 0)

	// Wait for HSI16 to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSI16RDYF) {
	}

	// Disable PLL
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)

	// Wait for PLL to be disabled
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL to use HSI16 with a PLLDIV of 2 and PLLMUL of 4.
	stm32.RCC.CFGR.ReplaceBits(
		(stm32.RCC_CFGR_PLLSRC_HSI16<<stm32.RCC_CFGR_PLLSRC_Pos)|
			(stm32.RCC_CFGR_PLLMUL_Mul4<<stm32.RCC_CFGR_PLLMUL_Pos)|
			(stm32.RCC_CFGR_PLLDIV_Div2<<stm32.RCC_CFGR_PLLDIV_Pos),
		stm32.RCC_CFGR_PLLSRC_Msk|
			stm32.RCC_CFGR_PLLMUL_Msk|
			stm32.RCC_CFGR_PLLDIV_Msk,
		0)

	// Enable PLL
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)

	// Wait for PLL to be ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Adjust flash latency
	if FlashLatency > getFlashLatency() {
		setFlashLatency(FlashLatency)
		for getFlashLatency() != FlashLatency {
		}
	}

	// HCLK
	stm32.RCC.CFGR.ReplaceBits(RCC_SYSCLK_DIV1, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Use PLL As System clock
	stm32.RCC.CFGR.ReplaceBits(stm32.RCC_CFGR_SWS_PLL, stm32.RCC_CFGR_SW_Msk, 0)
	for stm32.RCC.CFGR.Get()&stm32.RCC_CFGR_SW_Msk != stm32.RCC_CFGR_SWS_PLL {
	}

	// Set prescalers so half system clock (PCLKx = HCLK/2)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE1_Div2 << stm32.RCC_CFGR_PPRE1_Pos)
	stm32.RCC.CFGR.SetBits(stm32.RCC_CFGR_PPRE2_Div2 << stm32.RCC_CFGR_PPRE2_Pos)
}

func getFlashLatency() uint32 {
	return stm32.FLASH.ACR.Get() & stm32.Flash_ACR_LATENCY_Msk
}

func setFlashLatency(l uint32) {
	stm32.FLASH.ACR.ReplaceBits(l, stm32.Flash_ACR_LATENCY_Msk, 0)
}
