//go:build stm32u585

package runtime

import (
	"machine"
)

func init() {
	initCLK()

	initStdio()

	initTickTimer(&machine.TIM16)
}
