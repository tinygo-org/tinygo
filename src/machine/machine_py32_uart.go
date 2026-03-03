//go:build py32

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

// UART implements a minimal USART1 driver for PY32 parts.
type UART struct {
	Bus    *py32.USART_Type
	Buffer *RingBuffer
	irq    interrupt.Interrupt
}

var DefaultUART = &UART{Bus: py32.USART1, Buffer: NewRingBuffer()}

func (uart *UART) Configure(config UARTConfig) error {

	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Enable peripheral clock.
	py32.RCC.APBENR2.SetBits(py32.RCC_APBENR2_USART1EN)

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

	// Hook interrupt.
	uart.irq = interrupt.New(py32.IRQ_USART1, handleUartInterrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()

	configureDefaultUARTPins()

	return nil
}

// Configure pin for use by UART
func ConfigureUARTPin(pin Pin, af uint8) {
	pin.Configure(PinConfig{Mode: PinAlternate})
	pin.SetAltFunc(af)
}

func handleUartInterrupt(interrupt.Interrupt) {
	uart := DefaultUART
	data := uint8(uart.Bus.DR.Get())
	uart.Receive(data)
}

func (uart *UART) writeByte(c byte) error {
	for uart.Bus.SR.Get()&py32.USART_SR_TXE == 0 {
	}
	uart.Bus.DR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	for uart.Bus.SR.Get()&py32.USART_SR_TC == 0 {
	}
}
