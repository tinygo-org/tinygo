//go:build py32 && py32_uart_type

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

type UART struct {
	Bus    *py32.UART_Type
	Buffer *RingBuffer
	irq    interrupt.Interrupt
}

var DefaultUART = &UART{Bus: py32.UART1, Buffer: NewRingBuffer()}

func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	py32.RCC.APBENR1.SetBits(py32.RCC_APBENR1_UART1EN)
	uart.Bus.CR1.Set(0)
	uart.Bus.CR2.Set(py32.UART_CR2_RXNEIE)
	uart.Bus.CR3.Set(0)
	uart.Bus.BRR.Set((CPUFrequency() + config.BaudRate/2) / config.BaudRate)

	uart.irq = interrupt.New(py32.IRQ_UART1, handleUART1Interrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()
	configureDefaultUARTPins()
	return nil
}

func handleUART1Interrupt(interrupt.Interrupt) {
	DefaultUART.Receive(uint8(DefaultUART.Bus.DR.Get()))
}

func (uart *UART) writeByte(c byte) error {
	retries := uartTXRetries
	for retries > 0 && uart.Bus.SR.Get()&py32.UART_SR_TXE == 0 {
		retries--
	}
	if retries <= 0 {
		return errUARTWriteTimeout
	}
	uart.Bus.DR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	retries := uartTXRetries
	for retries > 0 && uart.Bus.SR.Get()&py32.UART_SR_TXE == 0 {
		retries--
	}
}
