// +build stm32

package machine

// Peripheral abstraction layer for UARTs on the stm32 family.

import (
	"device/stm32"
	"runtime/interrupt"
	"unsafe"
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Set the GPIO pins to defaults if they're not set
	if config.TX == 0 && config.RX == 0 {
		config.TX = UART_TX_PIN
		config.RX = UART_RX_PIN
	}

	// Enable USART clock
	enableAltFuncClock(unsafe.Pointer(uart.Bus))

	uart.configurePins(config)

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// Enable USART port, tx, rx and rx interrupts
	uart.Bus.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE)

	// Enable RX IRQ
	uart.Interrupt.SetPriority(0xc0)
	uart.Interrupt.Enable()
}

// handleInterrupt should be called from the appropriate interrupt handler for
// this UART instance.
func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	uart.Receive(byte((uart.Bus.DR.Get() & 0xFF)))
}

// SetBaudRate sets the communication speed for the UART. Defer to chip-specific
// routines for calculation
func (uart UART) SetBaudRate(br uint32) {
	divider := uart.getBaudRateDivisor(br)
	uart.Bus.BRR.Set(divider)
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	uart.Bus.DR.Set(uint32(c))

	for !uart.Bus.SR.HasBits(stm32.USART_SR_TXE) {
	}
	return nil
}
