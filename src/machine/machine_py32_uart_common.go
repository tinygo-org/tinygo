//go:build py32

package machine

import "runtime/interrupt"

type uartError string

func (e uartError) Error() string { return string(e) }

// errUARTWriteTimeout is returned when the transmit register stays busy.
const errUARTWriteTimeout uartError = "UART: write timeout"

const (
	errUARTInvalidTXPin uartError = "UART: invalid TX pin"
	errUARTInvalidRXPin uartError = "UART: invalid RX pin"
	errUARTPinsEqual    uartError = "UART: TX and RX pins must differ"
)

func configureUARTPins(uartNum uint8, config UARTConfig) error {
	if config.TX == 0 && config.RX == 0 {
		config.TX, config.RX = defaultUARTPins()
	}
	if config.TX == config.RX && config.TX != NoPin {
		return errUARTPinsEqual
	}
	if config.TX != NoPin {
		af, ok := uartPinAF(uartNum, config.TX, true)
		if !ok {
			return errUARTInvalidTXPin
		}
		config.TX.Configure(PinConfig{Mode: PinAlternate})
		config.TX.SetAltFunc(af)
	}
	if config.RX != NoPin {
		af, ok := uartPinAF(uartNum, config.RX, false)
		if !ok {
			return errUARTInvalidRXPin
		}
		config.RX.Configure(PinConfig{Mode: PinAlternate})
		config.RX.SetAltFunc(af)
	}
	return nil
}

func uartTXRetryBudget(baudRate uint32) uint32 {
	retries := CPUFrequency() / baudRate * 10
	if retries < 10 {
		return 10
	}
	return retries
}

func uartYield() {
	if !interrupt.In() {
		gosched()
	}
}
