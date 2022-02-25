//go:build stm32l5
// +build stm32l5

package machine

import (
	"device/stm32"
	"runtime/volatile"
)

func getEXTIConfigRegister(pin uint8) *volatile.Register32 {
	switch (pin & 0xf) / 4 {
	case 0:
		return &stm32.EXTI.EXTICR1
	case 1:
		return &stm32.EXTI.EXTICR2
	case 2:
		return &stm32.EXTI.EXTICR3
	case 3:
		return &stm32.EXTI.EXTICR4
	}
	return nil
}

func enableEXTIConfigRegisters() {
	// No-op
}
