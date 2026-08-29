//go:build py32 && !py32_uart_type && py32_uart_no_interrupt

package machine

// The F410 SVD does not provide USART interrupt metadata. Keep transmit and
// configuration support without inventing an IRQ number.
func configureUSART1Interrupt(uart *UART) {
}
