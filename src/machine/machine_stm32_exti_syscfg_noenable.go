//go:build stm32wlx
// +build stm32wlx

package machine

import (
	"runtime/volatile"

	"tinygo.org/x/device/stm32"
)

func getEXTIConfigRegister(pin uint8) *volatile.Register32 {
	switch (pin & 0xf) / 4 {
	case 0:
		return &stm32.SYSCFG.EXTICR1
	case 1:
		return &stm32.SYSCFG.EXTICR2
	case 2:
		return &stm32.SYSCFG.EXTICR3
	case 3:
		return &stm32.SYSCFG.EXTICR4
	}
	return nil
}

func enableEXTIConfigRegisters() {
	// No registers to enable
}
