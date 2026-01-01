//go:build stm32g0b1

package runtime

import (
	"machine"
)

func init() {
	initCLK()

	machine.InitSerial()

	initTickTimer(&machine.TIM3)
}
