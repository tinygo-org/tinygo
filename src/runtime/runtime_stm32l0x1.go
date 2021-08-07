// +build stm32l0x1

package runtime

import (
	"github.com/sago35/device/stm32"
	"machine"
)

const (
	FlashLatency = stm32.Flash_ACR_LATENCY_WS1
)

func init() {
	initCLK()

	machine.Serial.Configure(machine.UARTConfig{})

	initTickTimer(&machine.TIM21)
}
