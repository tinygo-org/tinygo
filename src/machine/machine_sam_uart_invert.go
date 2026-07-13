//go:build sam && (atsamd51 || atsame5x)

package machine

import "device/sam"

// Configure UART TX/RX line inversion using SVD-generated APIs.
func (uart *UART) setInversion(config UARTConfig) {
	if config.InvertTX {
		uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_INT_CTRLA_TXINV)
	} else {
		uart.Bus.CTRLA.ClearBits(sam.SERCOM_USART_INT_CTRLA_TXINV)
	}
	if config.InvertRX {
		uart.Bus.CTRLA.SetBits(sam.SERCOM_USART_INT_CTRLA_RXINV)
	} else {
		uart.Bus.CTRLA.ClearBits(sam.SERCOM_USART_INT_CTRLA_RXINV)
	}
}
