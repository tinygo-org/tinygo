//go:build stm32l0x1
// +build stm32l0x1

package runtime

import (
	"machine"
	"tinygo.org/x/device/stm32"
)

const (
	FlashLatency = stm32.Flash_ACR_LATENCY_WS1
)

func init() {
	initCLK()

	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM21)
}
