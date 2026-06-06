//go:build stm32 && !stm32u585

package machine

// SetBaudRate sets the communication speed for the UART. Defers to
// chip-specific getBaudRateDivisor for the divisor calculation.
//
// On STM32U585 this function is overridden in machine_stm32u585.go because
// the U5 family requires UE=0 to write BRR.
func (uart *UART) SetBaudRate(br uint32) {
	divider := uart.getBaudRateDivisor(br)
	uart.Bus.BRR.Set(divider)
}
