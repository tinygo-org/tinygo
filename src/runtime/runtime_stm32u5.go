//go:build stm32u5

package runtime

import (
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
	// This matches the known-working bare-metal C configuration for
	// the Arduino Uno Q (STM32U585). The MCU boots with MSI at 4MHz,
	// VOS Range 4, and 0 flash wait states. No additional configuration
	// is needed.
}
