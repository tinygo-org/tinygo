//go:build py32 && !py32_uart_type && !py32_uart_no_interrupt

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

func configureUSART1Interrupt(uart *UART) {
	uart.irq = interrupt.New(py32.IRQ_USART1, handleUSART1Interrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()
}
