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
	// Use MSI at 4MHz — the reset default clock configuration.
	// The MCU boots with MSI at 4MHz, VOS Range 4, and 0 flash wait states.

	// Enable PWR peripheral clock (required on STM32U5 before accessing PWR registers).
	stm32.RCC.AHB3ENR.SetBits(stm32.RCC_AHB3ENR_PWREN)
	_ = stm32.RCC.AHB3ENR.Get() // read-back for clock stabilization

	// Switch from VOS Range 4 to Range 3. Range 4 doesn't support ADC (RM0456 §6.3.6).
	// Range 3 supports up to 50 MHz HCLK and enables ADC. No flash wait-state change needed at 4 MHz.
	// The EPOD booster must be enabled before changing VOS (RM0456 §10.5.4).
	stm32.PWR.VOSR.SetBits(stm32.PWR_VOSR_BOOSTEN)
	stm32.PWR.VOSR.ReplaceBits(stm32.PWR_VOSR_VOS_Range3<<stm32.PWR_VOSR_VOS_Pos, stm32.PWR_VOSR_VOS_Msk, 0)
	for !stm32.PWR.VOSR.HasBits(stm32.PWR_VOSR_VOSRDY) {
	}
}
