//go:build stm32l0x2
// +build stm32l0x2

package runtime

import (
	"device/stm32"
	"machine"
)

const (
	FlashLatency = stm32.Flash_ACR_LATENCY_WS1
)

func init() {
	initCLK()

	machine.InitSerial()
	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM3)
}
