//go:build py32

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

// Remember the clock used for baud rate calculations so Configure() can be
// called without explicitly passing the clock.
var py32UARTClockHz uint32 = 24_000_000

// UART implements a minimal USART1 driver for PY32 parts.
type UART struct {
	Bus    *py32.USART_Type
	Buffer *RingBuffer
	irq    interrupt.Interrupt
}

var DefaultUART = &UART{Bus: py32.USART1, Buffer: NewRingBuffer()}

// ConfigureWithClock initializes the UART using the provided peripheral clock
// frequency (in Hz). This avoids assuming a fixed MCU clock.
func (uart *UART) ConfigureWithClock(config UARTConfig, clockHz uint32) error {
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	// Configure default pins if they weren't provided.
	if config.TX == 0 {
		ConfigureUARTPin(DEFAULT_UART_TX_PIN, DEFAULT_UART_TX_PIN_AF)
	}

	if config.RX == 0 {
		ConfigureUARTPin(DEFAULT_UART_RX_PIN, DEFAULT_UART_RX_PIN_AF)
	}

	// Enable peripheral clock.
	py32.RCC.APBENR2.SetBits(py32.RCC_APBENR2_USART1EN)

	// Reset control registers to a known state.
	uart.Bus.CR1.Set(0)
	uart.Bus.CR2.Set(0)
	uart.Bus.CR3.Set(0)

	// Oversampling by 16: BRR expects fck/baud.
	divider := (clockHz + (config.BaudRate / 2)) / config.BaudRate
	uart.Bus.BRR.Set(divider)

	// Enable transmitter, receiver, RX interrupt, and the peripheral.
	uart.Bus.CR1.Set(py32.USART_CR1_TE | py32.USART_CR1_RE | py32.USART_CR1_RXNEIE | py32.USART_CR1_UE)

	// Hook interrupt.
	uart.irq = interrupt.New(py32.IRQ_USART1, handleUartInterrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()

	return nil
}

// Configure uses the last stored clock (defaulting to 24 MHz). Call
// ConfigureWithClock for explicit control.
func (uart *UART) Configure(config UARTConfig) error {
	return uart.ConfigureWithClock(config, py32UARTClockHz)
}

// InitSerialWithClock configures the default Serial using the supplied
// peripheral clock frequency.
func InitSerialWithClock(clockHz uint32) {
	py32UARTClockHz = clockHz
	//Serial.ConfigureWithClock(UARTConfig{}, clockHz)
}

// Configure pin for use by UART
func ConfigureUARTPin(pin Pin, af uint8) {
	pin.enableClock()
	port, n := pin.getPort()
	pos := (n % 16) * 2

	// Alternate function mode is encoded as 0b10.
	port.MODER.ReplaceBits(2, gpioModeMask, pos)
	port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
	port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
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
