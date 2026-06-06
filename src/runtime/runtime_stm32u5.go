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
	// Configure SYSCLK to 160 MHz via PLL1 using MSIS (4 MHz reset default) as the source.
	// Formula: Fout = Fin × N / M / R = 4 × 80 / 1 / 2 = 160 MHz
	//   - M = 1 (Div1), N = 80 (stored as N-1 = 79), R = 2 (Div2)
	//   - VCO input = 4 MHz (PLL1RGE Range1: 4–8 MHz), VCO output = 320 MHz
	// VOS Range 1 (1.2V) is required for SYSCLK > 100 MHz (RM0456 §6.3.6).

	// Enable PWR peripheral clock (required on STM32U5 before accessing PWR registers).
	stm32.RCC.AHB3ENR.SetBits(stm32.RCC_AHB3ENR_PWREN)
	_ = stm32.RCC.AHB3ENR.Get() // read-back for clock stabilization

	// Enable the EPOD booster before raising VOS (RM0456 §10.5.4).
	stm32.PWR.VOSR.SetBits(stm32.PWR_VOSR_BOOSTEN)

	// Raise voltage scaling to Range 1 (1.2V) to support 160 MHz operation.
	stm32.PWR.VOSR.ReplaceBits(stm32.PWR_VOSR_VOS_Range1<<stm32.PWR_VOSR_VOS_Pos, stm32.PWR_VOSR_VOS_Msk, 0)

	// Wait for both the VOS regulator and the EPOD booster to become ready.
	for !stm32.PWR.VOSR.HasBits(stm32.PWR_VOSR_VOSRDY | stm32.PWR_VOSR_BOOSTRDY) {
	}

	// Set Flash latency to 4 wait states and enable prefetch before raising the clock
	// (required for 160 MHz at VOS Range 1, RM0456 §7.3.3).
	stm32.FLASH.ACR.ReplaceBits(4, stm32.Flash_ACR_LATENCY_Msk, 0)
	stm32.FLASH.ACR.SetBits(stm32.Flash_ACR_PRFTEN)

	// Configure PLL1: source = MSIS (4 MHz), M = 1, PLL1RGE = 4–8 MHz, R output enabled.
	stm32.RCC.PLL1CFGR.Set(
		stm32.RCC_PLL1CFGR_PLL1SRC_MSIS |
			(stm32.RCC_PLL1CFGR_PLL1RGE_Range1 << stm32.RCC_PLL1CFGR_PLL1RGE_Pos) |
			(stm32.RCC_PLL1CFGR_PLL1M_Div1 << stm32.RCC_PLL1CFGR_PLL1M_Pos) |
			stm32.RCC_PLL1CFGR_PLL1REN,
	)

	// Set PLL1 dividers: N = 80 (stored as N-1 = 79), R = 2 (Div2 = 1).
	stm32.RCC.PLL1DIVR.Set(
		(79 << stm32.RCC_PLL1DIVR_PLL1N_Pos) |
			(stm32.RCC_PLL1DIVR_PLL1R_Div2 << stm32.RCC_PLL1DIVR_PLL1R_Pos),
	)

	// Enable PLL1 and wait for it to lock.
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLL1ON)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLL1RDY) {
	}

	// Switch SYSCLK to PLL1 and wait for the hardware to confirm the switch.
	stm32.RCC.SetCFGR1_SW(stm32.RCC_CFGR1_SW_PLL)
	for stm32.RCC.GetCFGR1_SWS() != stm32.RCC_CFGR1_SWS_PLL {
	}
}
