//go:build py32 && !py32_uart_type && !py32_uart_clock_apb2 && !py32_usart1_clock_literal

package machine

import "device/py32"

func enableUSART1Clock() {
	py32.RCC.APBENR2.SetBits(py32.RCC_APBENR2_USART1EN)
}
