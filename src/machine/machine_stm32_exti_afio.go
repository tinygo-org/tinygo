//go:build stm32f1

package machine

import (
	"device/stm32"
	"runtime/volatile"
)

func getEXTIConfigRegister(pin uint8) *volatile.Register32 {
	switch (pin & 0xf) / 4 {
	case 0:
		return &stm32.AFIO.EXTICR1
	case 1:
		return &stm32.AFIO.EXTICR2
	case 2:
		return &stm32.AFIO.EXTICR3
	case 3:
		return &stm32.AFIO.EXTICR4
	}
	return nil
}

func enableEXTIConfigRegisters() {
	// Enable AFIO
	stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
}
