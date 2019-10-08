// +build fe310

package machine

import (
	"device/sifive"
)

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	sifive.GPIO0.INPUT_EN.SetBits(1 << uint8(p))
	if config.Mode == PinOutput {
		sifive.GPIO0.OUTPUT_EN.SetBits(1 << uint8(p))
	}
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	if high {
		sifive.GPIO0.PORT.SetBits(1 << uint8(p))
	} else {
		sifive.GPIO0.PORT.ClearBits(1 << uint8(p))
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	val := sifive.GPIO0.VALUE.Get() & (1 << uint8(p))
	println(sifive.GPIO0.VALUE.Get())
	return (val > 0)
}

type UART struct {
	Bus    *sifive.UART_Type
	Buffer *RingBuffer
}

var (
	UART0 = UART{Bus: sifive.UART0, Buffer: NewRingBuffer()}
)

func (uart UART) Configure(config UARTConfig) {
	// Assuming a 16Mhz Crystal (which is Y1 on the HiFive1), the divisor for a
	// 115200 baud rate is 138.
	sifive.UART0.DIV.Set(138)
	sifive.UART0.TXCTRL.Set(sifive.UART_TXCTRL_ENABLE)
}

func (uart UART) WriteByte(c byte) {
	for sifive.UART0.TXDATA.Get()&sifive.UART_TXDATA_FULL != 0 {
	}

	sifive.UART0.TXDATA.Set(uint32(c))
}
