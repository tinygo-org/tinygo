//go:build stm32 && (stm32f1 || stm32f4)

package machine

// Hardware inversion is not supported on F1/F4.
func (uart *UART) setInversion(config UARTConfig) {
}
