//go:build sam && (atsamd51 || atsame5x)

package machine

// Configure UART TX/RX line inversion using SVD-generated APIs.
func (uart *UART) setInversion(config UARTConfig) {
	if config.InvertTX {
		uart.Bus.SetCTRLA_TXINV(1)
	} else {
		uart.Bus.SetCTRLA_TXINV(0)
	}
	if config.InvertRX {
		uart.Bus.SetCTRLA_RXINV(1)
	} else {
		uart.Bus.SetCTRLA_RXINV(0)
	}
}
