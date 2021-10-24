// +build stm32,!stm32f1,!stm32l5

package machine

import (
	"device/stm32"
	"runtime/volatile"
)

func getEXTIConfigRegister(pin uint8) *volatile.Register32 {
	switch (pin & 0xf) / 4 {
	case 0:
		return &stm32.SYSCFG.EXTICR1.Register32
	case 1:
		return &stm32.SYSCFG.EXTICR2.Register32
	case 2:
		return &stm32.SYSCFG.EXTICR3.Register32
	case 3:
		return &stm32.SYSCFG.EXTICR4.Register32
	}
	return nil
}

func enableEXTIConfigRegisters() {
	// Enable SYSCFG
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
}
