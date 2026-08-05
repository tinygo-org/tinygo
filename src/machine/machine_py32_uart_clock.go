//go:build py32 && !py32_uart_type && !py32_uart_clock_apb2

package machine

import "device/py32"

func enableUSART1Clock() {
	// USART1EN is bit 14 even in SVDs that omit the field metadata.
	py32.RCC.APBENR2.SetBits(1 << 14)
}
