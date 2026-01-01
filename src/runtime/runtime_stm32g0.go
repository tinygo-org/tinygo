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
	// Initialize clock to use HSI16 at 16MHz (matching proven HAL configuration)
	// The STM32G0 starts up with HSI16 (16MHz internal RC) as the system clock

	// Enable PWR clock
	stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_PWREN)
	// Read back to ensure the write is complete (memory barrier)
	_ = stm32.RCC.APBENR1.Get()

	// Set Power Regulator to enable max performance (Range 1)
	// VOS = 01 for Range 1 (high performance, up to 64 MHz)
	stm32.PWR.CR1.ReplaceBits(1<<stm32.PWR_CR1_VOS_Pos, stm32.PWR_CR1_VOS_Msk, 0)

	// Wait for voltage scaling to be ready (VOSF = 0 means ready)
	for stm32.PWR.SR2.HasBits(stm32.PWR_SR2_VOSF) {
	}

	// Enable HSI16 (should already be on after reset, but ensure it)
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Set HSI16 division factor to 1 (no division) - HSIDIV = 000
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_HSIDIV_Msk)

	// For 16 MHz operation, flash latency can be 0 wait states (up to 24 MHz in Range 1)
	stm32.FLASH.ACR.ReplaceBits(0, stm32.Flash_ACR_LATENCY_Msk, 0)
	for (stm32.FLASH.ACR.Get() & stm32.Flash_ACR_LATENCY_Msk) != 0 {
	}

	// Set AHB prescaler to 1 (no division) - HPRE = 0xxx
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_HPRE_Msk, 0)

	// Set APB prescaler to 1 (no division) - PPRE = 0xx
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_PPRE_Msk, 0)

	// Select HSI16 as system clock source (SW = 000)
	// Note: After reset, HSI16 should already be the system clock source
	stm32.RCC.CFGR.ReplaceBits(0, stm32.RCC_CFGR_SW_Msk, 0)
	for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != 0 {
	}
}
