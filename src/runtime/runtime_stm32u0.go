//go:build stm32u0

package runtime

import "machine"

func init() {
	initStdio()
	initTickTimer(&machine.TIM16)
}
