//go:build stm32 && stm32f7x2
// +build stm32,stm32f7x2

package runtime

import (
	"device/stm32"
	"machine"
)

/*
clock settings

	+-------------+--------+
	| HSE         | 8mhz   |
	| SYSCLK      | 216mhz |
	| HCLK        | 216mhz |
	| APB1(PCLK1) | 27mhz  |
	| APB2(PCLK2) | 108mhz |
	+-------------+--------+
*/
const (
	HSE_STARTUP_TIMEOUT = 0x0500
	PLL_M               = 4
	PLL_N               = 216
	PLL_P               = 2
	PLL_Q               = 2
)

func init() {
	initCLK()

	machine.InitSerial()

	initTickTimer(&machine.TIM3)
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

func initCLK() {
	// PWR_CLK_ENABLE
	stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	_ = stm32.RCC.APB1ENR.Get()

	// PWR_VOLTAGESCALING_CONFIG
	stm32.PWR.CR1.ReplaceBits(0x3<<stm32.PWR_CR1_VOS_Pos, stm32.PWR_CR1_VOS_Msk, 0)
	_ = stm32.PWR.CR1.Get()

	// Initialize the High-Speed External Oscillator
	initOsc()

	// Set flash wait states (min 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) < 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// HCLK (0x1C00 = DIV_16, 0x0 = RCC_SYSCLK_DIV1) - ensure timers remain
	// within spec as the SYSCLK source changes.
	stm32.RCC.CFGR.ReplaceBits(0x00001C00, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x00001C00<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set SYSCLK source and wait
	// (2 = PLLCLK, 3 = RCC_CFGR_SW mask, 3 << 3 = RCC_CFGR_SWS mask)
	stm32.RCC.CFGR.ReplaceBits(2, 3, 0)
	for stm32.RCC.CFGR.Get()&(3<<2) != (2 << 2) {
	}

	// Set flash wait states (max 7 latency units) based on clock
	if (stm32.FLASH.ACR.Get() & stm32.FLASH_ACR_LATENCY_Msk) > 7 {
		stm32.FLASH.ACR.ReplaceBits(7, stm32.FLASH_ACR_LATENCY_Msk, 0)
	}

	// Set APB1 and APB2 clocks (0x1800 = DIV8, 0x1000 = DIV2)
	stm32.RCC.CFGR.ReplaceBits(0x1800, stm32.RCC_CFGR_PPRE1_Msk, 0)
	stm32.RCC.CFGR.ReplaceBits(0x1000<<3, stm32.RCC_CFGR_PPRE2_Msk, 0)
}

func initOsc() {
	// Enable HSE, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSEON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSERDY) {
	}

	// Disable the PLL, wait until disabled
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLLON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure the PLL
	stm32.RCC.PLLCFGR.Set(0x20000000 |
		(1 << stm32.RCC_PLLCFGR_PLLSRC_Pos) | // 1 = HSE
		PLL_M |
		(PLL_N << stm32.RCC_PLLCFGR_PLLN_Pos) |
		(((PLL_P >> 1) - 1) << stm32.RCC_PLLCFGR_PLLP_Pos) |
		(PLL_Q << stm32.RCC_PLLCFGR_PLLQ_Pos))

	// Enable the PLL, wait until ready
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}
}
