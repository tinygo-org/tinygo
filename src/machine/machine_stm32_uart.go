// +build stm32

package machine

// Peripheral abstraction layer for UARTs on the stm32 family.

import (
	"device/arm"
	"device/stm32"
	"unsafe"
)

// UART representation
type UART struct {
	Buffer *RingBuffer
	Bus    *stm32.USART_Type
	IRQVal uint32
}

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Set the GPIO pins to defaults if they're not set
	if config.TX == 0 {
		config.TX = UART_TX_PIN
	}
	if config.RX == 0 {
		config.RX = UART_RX_PIN
	}
	uart.configurePins(config)

	// Enable USART clock
	EnableAltFuncClock(unsafe.Pointer(uart.Bus))

	// Set baud rate
	uart.SetBaudRate(config.BaudRate)

	// Enable USART port, tx, rx and rx interrupts
	uart.Bus.CR1.Set(stm32.USART_CR1_TE | stm32.USART_CR1_RE | stm32.USART_CR1_RXNEIE | stm32.USART_CR1_UE)

	// Enable RX IRQ
	arm.SetPriority(uart.IRQVal, 0xc0)
	arm.EnableIRQ(uart.IRQVal)
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
