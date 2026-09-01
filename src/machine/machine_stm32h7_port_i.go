//go:build stm32 && stm32h7 && !stm32h723

package machine

import (
	"device/stm32"
)

func pinGetPortI() *stm32.GPIO_Type {
	return stm32.GPIOI
}

func (p Pin) enableClockOther() {
	switch p.getPort() {
	case stm32.GPIOI:
		stm32.RCC.AHB4ENR.SetBits(stm32.RCC_AHB4ENR_GPIOIEN)
	}
}
