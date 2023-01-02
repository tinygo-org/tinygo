//go:build (nrf52840 || nrf52833 || nrf9160) && !nrf9160

package machine

import (
	"device/nrf"
	"runtime/interrupt"
)

// UART
var (
	// UART1 is the hardware EasyDMA UART on the NRF SoC.
	UART1  = &_UART1
	_UART1 = UARTE{UARTE_Type: nrf.UARTE1, Buffer: NewRingBuffer()}
)

func init() {
	UART1.Interrupt = interrupt.New(nrf.IRQ_UARTE1, _UART1.handleInterrupt)
}
