//go:build stm32 && stm32h7 && stm32h757_cm7

package machine

import (
	"device/stm32"
	"runtime/volatile"
)

func intrMaskReg() *volatile.Register32 {
	return &stm32.EXTI.C1IMR1
}

func intrPendReg() *volatile.Register32 {
	return &stm32.EXTI.C1PR1
}
