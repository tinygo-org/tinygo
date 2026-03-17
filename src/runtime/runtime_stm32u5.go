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
}
