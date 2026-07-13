//go:build sam && !atsamd51 && !atsame5x

package machine

// Hardware inversion is not supported on SAMD21.
func (uart *UART) setInversion(config UARTConfig) {
}
