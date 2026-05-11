//go:build stm32u0

package runtime

import "machine"

func init() {
	machine.InitSerial()
	initTickTimer(&machine.TIM16)
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}
