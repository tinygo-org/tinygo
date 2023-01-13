//go:build atmega || esp || nrf || sam || sifive || stm32 || k210 || nxp || rp2040
// +build atmega esp nrf sam sifive stm32 k210 nxp rp2040

package machine

import "errors"

var errUARTBufferEmpty = errors.New("UART buffer empty")

// UARTParity is the parity setting to be used for UART communication.
type UARTParity uint8

const (
	// ParityNone means to not use any parity checking. This is
	// the most common setting.
	ParityNone UARTParity = iota

	// ParityEven means to expect that the total number of 1 bits sent
	// should be an even number.
	ParityEven

	// ParityOdd means to expect that the total number of 1 bits sent
	// should be an odd number.
	ParityOdd
)
