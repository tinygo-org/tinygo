//go:build fu540

package machine

import (
	"device/sifive"
	"errors"
	"runtime/interrupt"
)

const deviceName = sifive.Device

func CPUFrequency() uint32 {
	return 390000000
}

var (
	errUnsupportedSPIController = errors.New("SPI controller not supported. Use SPI0 or SPI1.")
	errI2CTxAbort               = errors.New("I2C transmition has been aborted.")
)

type UART struct {
	Bus    *sifive.UART_Type
	Buffer *RingBuffer
}

var (
	DefaultUART = UART0
	UART0       = &_UART0
	_UART0      = UART{Bus: sifive.UART0, Buffer: NewRingBuffer()}
)

func (uart *UART) Configure(config UARTConfig) {
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
}

func (uart *UART) writeByte(c byte) error {
	return nil
}

func (uart *UART) flush() {}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	panic("Set")
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	panic("Get")
	return false
}

// TODO: SPI
