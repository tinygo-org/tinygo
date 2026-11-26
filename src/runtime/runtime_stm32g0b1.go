//go:build stm32g0b1

package runtime

import (
	"machine"
)

func init() {
	initCLK()

	initStdio()

	initTickTimer(&machine.TIM3)
}
