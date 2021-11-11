//go:build stm32wle5
// +build stm32wle5

package runtime

import (
	"machine"
)

type arrtype = uint32

func init() {
	// Currently the clock is not configured, which means
	// the MCU runs at default reset clock speed (4Mhz).
	// Code to initialize RCC and PLL can go here.

	// UART init
	machine.Serial.Configure(machine.UARTConfig{})

	// Timers init
	initTickTimer(&machine.TIM1)
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}
