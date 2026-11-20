//go:build (rp2040 || rp2350) && machine.uart && !boardpackage

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}

	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART1,
	}
)

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.HandleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.HandleInterrupt)
}

func defaultUARTPins(config *UARTConfig) error {
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}
	return nil
}
