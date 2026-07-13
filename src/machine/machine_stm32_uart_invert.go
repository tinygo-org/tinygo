//go:build stm32 && !(stm32f1 || stm32f4)

package machine

// Configure UART TX/RX line inversion using SVD-generated APIs.
func (uart *UART) setInversion(config UARTConfig) {
	if config.InvertTX {
		uart.Bus.SetCR2_TXINV(1)
	} else {
		uart.Bus.SetCR2_TXINV(0)
	}
	if config.InvertRX {
		uart.Bus.SetCR2_RXINV(1)
	} else {
		uart.Bus.SetCR2_RXINV(0)
	}
}
