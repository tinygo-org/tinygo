//go:build stm32u585

package runtime

import (
	"machine"
)

func init() {
	initCLK()

	machine.InitSerial()

	initTickTimer(&machine.TIM16)
}
