//go:build stm32 && stm32h7 && stm32h723

package machine

import (
	"device/stm32"
)

func pinGetPortI() *stm32.GPIO_Type {
	panic("machine: unknown port")
}

func (p Pin) enableClockOther() {}
