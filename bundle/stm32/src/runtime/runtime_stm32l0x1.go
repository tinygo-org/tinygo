//go:build stm32l0x1

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

	initTickTimer(&machine.TIM21)
}
