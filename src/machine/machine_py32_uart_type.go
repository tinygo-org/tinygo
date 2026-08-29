//go:build py32 && py32_uart_type

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

type UART struct {
	Bus       *py32.UART_Type
	Buffer    *RingBuffer
	irq       interrupt.Interrupt
	txRetries uint32
	num       uint8
}

var DefaultUART = &UART{Bus: py32.UART1, Buffer: NewRingBuffer(), num: 1}

func (uart *UART) Configure(config UARTConfig) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	uart.txRetries = uartTXRetryBudget(config.BaudRate)
	if err := configureUARTPins(uart.num, config); err != nil {
		return err
	}

	py32.RCC.APBENR1.SetBits(py32.RCC_APBENR1_UART1EN)
	uart.Bus.CR1.Set(py32.UART_CR1_M_Char8Bits)
	uart.Bus.CR2.Set(py32.UART_CR2_RXNEIE)
	uart.Bus.CR3.Set(0)
	divider := (CPUFrequency() + config.BaudRate*8) / (config.BaudRate * 16)
	uart.Bus.BRR.Set(divider)

	uart.irq = interrupt.New(py32.IRQ_UART1, handleUART1Interrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()
	return nil
}

func handleUART1Interrupt(interrupt.Interrupt) {
	DefaultUART.Receive(uint8(DefaultUART.Bus.DR.Get()))
}

func (uart *UART) writeByte(c byte) error {
	retries := uart.txRetries
	for retries > 0 && uart.Bus.SR.Get()&py32.UART_SR_TDRE == 0 {
		uartYield()
		retries--
	}
	if retries <= 0 {
		return errUARTWriteTimeout
	}
	uart.Bus.DR.Set(uint32(c))
	return nil
}

func (uart *UART) flush() {
	retries := uart.txRetries
	for retries > 0 && uart.Bus.SR.Get()&py32.UART_SR_TXE == 0 {
		uartYield()
		retries--
	}
}
