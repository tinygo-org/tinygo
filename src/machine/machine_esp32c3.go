// +build esp32c3

package machine

import (
	"device/esp"
	"runtime/volatile"
)

// CPUFrequency returns the current CPU frequency of the chip.
// Currently it is a fixed frequency but it may allow changing in the future.
func CPUFrequency() uint32 {
	return 160e6 // 160MHz
}

const (
	MODULE_GPIO = iota
	MODULE_UART0
	MODULE_UART1
)

func InterruptForModule(module int) uint32 {
	switch module {
	case MODULE_GPIO:
		return 6
	case MODULE_UART0, MODULE_UART1:
		return 7
	}
	return 0
}

func MapRegisterForModule(module int) *volatile.Register32 {
	switch module {
	case MODULE_GPIO:
		return &esp.INTERRUPT_CORE0.GPIO_INTERRUPT_PRO_MAP
	case MODULE_UART0:
		return &esp.INTERRUPT_CORE0.UART_INTR_MAP
	case MODULE_UART1:
		return &esp.INTERRUPT_CORE0.UART1_INTR_MAP
	}
	return nil
}
