//go:build py32 && !py32_uart_type && !py32_uart_clock_apb2 && py32_usart1_clock_literal

package machine

import "device/py32"

func enableUSART1Clock() {
	// The F001 SVDs omit USART1EN; the vendor register layout places it at bit 14.
	py32.RCC.APBENR2.SetBits(1 << 14)
}
