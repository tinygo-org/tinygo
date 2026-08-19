//go:build py32

package machine

import "runtime/interrupt"

// errUARTWriteTimeout is returned by writeByte when the TX register does not
// become empty within the retry budget.
type uartError string

func (e uartError) Error() string { return string(e) }

const errUARTWriteTimeout uartError = "UART: write timeout"

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
