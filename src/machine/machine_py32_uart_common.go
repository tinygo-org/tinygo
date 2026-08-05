//go:build py32

package machine

// errUARTWriteTimeout is returned by writeByte when the TX register does not
// become empty within the retry budget.
type uartError string

func (e uartError) Error() string { return string(e) }

const errUARTWriteTimeout uartError = "UART: write timeout"

// uartTXRetries is the upper bound on the transmit status polling loops. At
// 48 MHz, 10,000 peripheral reads cover more than one byte at 9600 baud.
const uartTXRetries = 10000
