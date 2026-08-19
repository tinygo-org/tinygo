//go:build py32 && !py32_uart_type

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

// UART implements a minimal USART driver for PY32 parts. It works with any of
// the on-chip USART peripherals: the peripheral-specific wiring (clock gate and
// RX interrupt) is provided by the setup function stored on the instance, which
// is set when the instance is created (see setupUSART1 / setupUSART2).
type UART struct {
	Bus       *py32.USART_Type
	Buffer    *RingBuffer
	irq       interrupt.Interrupt
	setup     func(*UART)
	txRetries uint32
}

// DefaultUART is the first USART (USART1) and backs machine.Serial.
var DefaultUART = &UART{Bus: defaultUSART(), Buffer: NewRingBuffer(), setup: setupUSART1}

func (uart *UART) Configure(config UARTConfig) error {

	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.txRetries = uartTXRetryBudget(config.BaudRate)

	// Enable the peripheral clock and hook its RX interrupt.
	uart.setup(uart)

	// Reset control registers to a known state.
	uart.Bus.CR1.Set(0)
	uart.Bus.CR2.Set(0)
	uart.Bus.CR3.Set(0)

	clockHz := CPUFrequency()

	// Oversampling by 16: BRR expects fck/baud.
	divider := (clockHz + (config.BaudRate / 2)) / config.BaudRate
	uart.Bus.BRR.Set(divider)

	// Enable transmitter, receiver, RX interrupt, and the peripheral.
	uart.Bus.CR1.Set(py32.USART_CR1_TE | py32.USART_CR1_RE | py32.USART_CR1_RXNEIE | py32.USART_CR1_UE)

	configureDefaultUARTPins()

	return nil
}

// setupUSART1 enables the USART1 peripheral clock and installs its RX interrupt
// handler. It records the instance in usart1RX so the handler can reach it
// without the var initializer forming an initialization cycle.
func setupUSART1(uart *UART) {
	usart1RX = uart
	enableUSART1Clock()
	configureUSART1Interrupt(uart)
}

func handleUSART1Interrupt(interrupt.Interrupt) {
	usart1RX.Receive(uint8(readUSARTData(usart1RX.Bus)))
}

// usart1RX is the UART whose RX interrupt is serviced by handleUSART1Interrupt.
// It is set by setupUSART1 at Configure time; keeping it as a plain package var
// (rather than referencing DefaultUART from the handler) avoids an
// initialization cycle so DefaultUART.setup can be set in its var initializer.
var usart1RX *UART

func (uart *UART) writeByte(c byte) error {
	retries := uart.txRetries
	for retries > 0 && !usartTXReady(uart.Bus) {
		uartYield()
		retries--
	}
	if retries <= 0 {
		return errUARTWriteTimeout
	}
	writeUSARTData(uart.Bus, uint32(c))
	return nil
}

func (uart *UART) flush() {
	retries := uart.txRetries
	for retries > 0 && !usartTXComplete(uart.Bus) {
		uartYield()
		retries--
	}
}
