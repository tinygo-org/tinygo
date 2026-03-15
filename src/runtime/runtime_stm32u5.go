//go:build stm32u5

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
	// Initialize clock to 160MHz using PLL1 with HSI16 as source.
	// PLL1 configuration: HSI16 (16MHz) / PLLM(1) * PLLN(10) / PLLR(1) = 160MHz
	// VCO = 16 * 10 = 160MHz, PLLR = /1 -> 160MHz SYSCLK

	// Enable PWR clock
	stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_CRSEN) // CRS enable not needed but ensures APB1 is clocked
	_ = stm32.RCC.APB1ENR1.Get()

	// Set voltage scaling to Range 1 for 160MHz operation
	// On U5, voltage scaling is in PWR.VOSR register
	// VOS = 11 (Range 1, up to 160MHz)
	stm32.PWR.VOSR.ReplaceBits(stm32.PWR_VOSR_VOS_Range1, stm32.PWR_VOSR_VOS_Msk, 0)
	// Wait for VOS ready
	for !stm32.PWR.VOSR.HasBits(stm32.PWR_VOSR_VOSRDY) {
	}

	// Enable EPOD booster for high performance (required for Range 1 >100MHz)
	stm32.PWR.VOSR.SetBits(stm32.PWR_VOSR_BOOSTEN)
	for !stm32.PWR.VOSR.HasBits(stm32.PWR_VOSR_BOOSTRDY) {
	}

	// Enable HSI16
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// Disable PLL1 before configuration
	stm32.RCC.CR.ClearBits(stm32.RCC_CR_PLL1ON)
	for stm32.RCC.CR.HasBits(stm32.RCC_CR_PLL1RDY) {
	}

	// Configure PLL1:
	// Source = HSI16 (0x2)
	// PLLM = 0 (divide by 1)
	// PLL1RGE = 0x3 (input range 8-16MHz)
	// Enable PLLR output
	stm32.RCC.PLL1CFGR.Set(
		(stm32.RCC_PLL1CFGR_PLL1SRC_HSI16 << stm32.RCC_PLL1CFGR_PLL1SRC_Pos) |
			(stm32.RCC_PLL1CFGR_PLL1M_Div1 << stm32.RCC_PLL1CFGR_PLL1M_Pos) |
			(stm32.RCC_PLL1CFGR_PLL1RGE_Range2 << stm32.RCC_PLL1CFGR_PLL1RGE_Pos))

	// Enable PLL1R output
	stm32.RCC.PLL1CFGR.SetBits(1 << stm32.RCC_PLL1CFGR_PLL1REN_Pos)

	// Set PLL1 dividers in PLL1DIVR:
	// PLL1N = 10 (value - 1 = 9 in register)
	// PLL1R = 1 (value - 1 = 0 in register)
	// VCO = HSI16/1 * 10 = 160MHz
	// PLLR output = 160MHz / 1 = 160MHz
	stm32.RCC.SetPLL1DIVR_PLL1N(9) // N = 10, register value = N-1 = 9
	stm32.RCC.SetPLL1DIVR_PLL1R(0) // R = 1, register value = R-1 = 0

	// Enable PLL1
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLL1ON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLL1RDY) {
	}

	// Set flash latency to 4 wait states (required for 160MHz in Range 1)
	const FLASH_LATENCY_4 = 4
	stm32.FLASH.ACR.ReplaceBits(FLASH_LATENCY_4, stm32.Flash_ACR_LATENCY_Msk, 0)
	for (stm32.FLASH.ACR.Get() & stm32.Flash_ACR_LATENCY_Msk) != FLASH_LATENCY_4 {
	}

	// Set AHB prescaler to 1 (no division) in CFGR2
	stm32.RCC.CFGR2.ReplaceBits(stm32.RCC_CFGR2_HPRE_Div1, stm32.RCC_CFGR2_HPRE_Msk, 0)

	// Set APB1 and APB2 prescalers to 1 (no division) in CFGR2
	stm32.RCC.CFGR2.ReplaceBits(stm32.RCC_CFGR2_PPRE1_Div1, stm32.RCC_CFGR2_PPRE1_Msk, 0)
	stm32.RCC.CFGR2.ReplaceBits(stm32.RCC_CFGR2_PPRE2_Div1, stm32.RCC_CFGR2_PPRE2_Msk, 0)

	// Switch system clock to PLL1 (SW = 11 = PLL)
	stm32.RCC.CFGR1.ReplaceBits(stm32.RCC_CFGR1_SW_PLL, stm32.RCC_CFGR1_SW_Msk, 0)
	for (stm32.RCC.CFGR1.Get() & stm32.RCC_CFGR1_SWS_Msk) != (stm32.RCC_CFGR1_SWS_PLL << stm32.RCC_CFGR1_SWS_Pos) {
	}
}
