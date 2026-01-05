//go:build stm32g0

package runtime

import (
	"device/stm32"
	"machine"
)

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
	// Initialize clock to 64MHz using PLL with HSI16 as source
	// PLL configuration: HSI16 (16MHz) / PLLM(1) * PLLN(8) / PLLR(2) = 64MHz

	// Enable PWR clock
	stm32.RCC.SetAPBENR1_PWREN(1)
	// Read back to ensure the write is complete (memory barrier)
	_ = stm32.RCC.APBENR1.Get()

	// Set Power Regulator to enable max performance (Range 1)
	// VOS = 01 for Range 1 (high performance, up to 64 MHz)
	stm32.PWR.SetCR1_VOS(1)
	// Wait for voltage scaling to be ready (VOSF = 0 means ready)
	for stm32.PWR.SR2.HasBits(stm32.PWR_SR2_VOSF) {
	}

	// Enable HSI16
	stm32.RCC.SetCR_HSION(1)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Set HSI16 division factor to 1 (no division) - HSIDIV = 000
	stm32.RCC.SetCR_HSIDIV(0)

	// Disable PLL before configuration
	stm32.RCC.SetCR_PLLON(0)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Configure PLL: HSI16 / 1 * 8 / 2 = 64 MHz
	// PLLSRC = HSI16 (2)
	// PLLM = 0 (divide by 1)
	// PLLN = 8 (multiply by 8) -> VCO = 16 * 8 = 128 MHz
	// PLLR = 0 (divide by 2) -> SYSCLK = 128 / 2 = 64 MHz
	// PLLREN = 1 (enable R output for SYSCLK)
	const (
		PLLSRC_HSI16 = 2 // HSI16 as PLL source
		PLLM_DIV1    = 0 // /1
		PLLN_MUL8    = 8 // *8
		PLLR_DIV2    = 0 // /2 (0 = divide by 2)
	)
	stm32.RCC.PLLCFGR.Set(
		(PLLSRC_HSI16 << stm32.RCC_PLLCFGR_PLLSRC_Pos) |
			(PLLM_DIV1 << stm32.RCC_PLLCFGR_PLLM_Pos) |
			(PLLN_MUL8 << stm32.RCC_PLLCFGR_PLLN_Pos) |
			(PLLR_DIV2 << stm32.RCC_PLLCFGR_PLLR_Pos) |
			stm32.RCC_PLLCFGR_PLLREN) // Enable PLLR output

	// Enable PLL
	stm32.RCC.SetCR_PLLON(1)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// Set flash latency to 2 wait states (required for 64MHz in Range 1)
	// Must be set BEFORE switching to higher frequency clock
	const FLASH_LATENCY_2 = 2
	stm32.FLASH.SetACR_LATENCY(FLASH_LATENCY_2)
	for (stm32.FLASH.ACR.Get() & stm32.Flash_ACR_LATENCY_Msk) != FLASH_LATENCY_2 {
	}

	// Set AHB prescaler to 1 (no division)
	stm32.RCC.SetCFGR_HPRE(0)
	// Set APB prescaler to 1 (no division)
	stm32.RCC.SetCFGR_PPRE(0)

	// Switch system clock to PLL (SW = 010)
	const RCC_CFGR_SW_PLL = 2
	stm32.RCC.SetCFGR_SW(RCC_CFGR_SW_PLL)
	// Wait for PLL to be used as system clock (SWS = 010)
	for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != (RCC_CFGR_SW_PLL << stm32.RCC_CFGR_SWS_Pos) {
	}
}
