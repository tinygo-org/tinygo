// +build k210

package machine

import (
	"device/kendryte"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 390000000
}

type PinMode uint8

const (
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
)

type fpioaPullMode uint8

const (
	fpioaPullNone fpioaPullMode = iota
	fpioaPullDown
	fpioaPullUp
)

func (p Pin) fpioaSetIOPull(pull fpioaPullMode) {
	switch pull {
	case fpioaPullNone:
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PU & kendryte.FPIOA_IO_PD)
	case fpioaPullUp:
		kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_PU)
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PD)
	case fpioaPullDown:
		kendryte.FPIOA.IO[uint8(p)].ClearBits(kendryte.FPIOA_IO_PU)
		kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_PD)
	}
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	var input bool

	kendryte.FPIOA.IO[uint8(p)].SetBits(kendryte.FPIOA_IO_OE_EN | kendryte.FPIOA_IO_IE_EN | kendryte.FPIOA_IO_ST | kendryte.FPIOA_IO_DS_Msk)
	switch config.Mode {
	case PinInput:
		p.fpioaSetIOPull(fpioaPullNone)
		input = true

	case PinInputPullUp:
		p.fpioaSetIOPull(fpioaPullUp)
		input = true

	case PinInputPullDown:
		p.fpioaSetIOPull(fpioaPullDown)
		input = true

	case PinOutput:
		p.fpioaSetIOPull(fpioaPullNone)
		input = false

	}

	if p >= P08 && p <= P15 {
		// Converts the IO pin number in the effective GPIO number (assuming default FPIOA function).
		gpioPin := uint8(p) - 8

		if input {
			kendryte.GPIO.DIRECTION.ClearBits(1 << gpioPin)
		} else {
			kendryte.GPIO.DIRECTION.SetBits(1 << gpioPin)
		}
	} else if p >= P16 && p <= P47 {
		// Converts the IO pin number in the effective GPIOHS number (assuming default FPIOA function).
		gpioPin := uint8(p) - 16

		if input {
			kendryte.GPIOHS.INPUT_EN.SetBits(1 << gpioPin)
			kendryte.GPIOHS.OUTPUT_EN.ClearBits(1 << gpioPin)
		} else {
			kendryte.GPIOHS.OUTPUT_EN.SetBits(1 << gpioPin)
			kendryte.GPIOHS.INPUT_EN.ClearBits(1 << gpioPin)
		}
	}
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	if p >= P08 && p <= P15 {
		gpioPin := uint8(p) - 8

		if high {
			kendryte.GPIO.DATA_OUTPUT.SetBits(1 << gpioPin)
		} else {
			kendryte.GPIO.DATA_OUTPUT.ClearBits(1 << gpioPin)
		}
	} else if p >= P16 && p <= P47 {
		gpioPin := uint8(p) - 16

		if high {
			kendryte.GPIOHS.OUTPUT_VAL.SetBits(1 << gpioPin)
		} else {
			kendryte.GPIOHS.OUTPUT_VAL.ClearBits(1 << gpioPin)
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	var val uint32
	if p >= P08 && p <= P15 {
		gpioPin := uint8(p) - 8
		val = kendryte.GPIO.DATA_INPUT.Get() & (1 << gpioPin)
	} else if p >= P16 && p <= P47 {
		gpioPin := uint8(p) - 16
		val = kendryte.GPIOHS.INPUT_VAL.Get() & (1 << gpioPin)
	}
	return (val > 0)
}

type UART struct {
	Bus    *kendryte.UARTHS_Type
	Buffer *RingBuffer
}

var (
	UART0 = UART{Bus: kendryte.UARTHS, Buffer: NewRingBuffer()}
)

func (uart UART) Configure(config UARTConfig) {
	div := CPUFrequency()/115200 - 1

	uart.Bus.DIV.Set(div)
	uart.Bus.TXCTRL.Set(kendryte.UARTHS_TXCTRL_TXEN)
	uart.Bus.RXCTRL.Set(kendryte.UARTHS_RXCTRL_RXEN)

	// Enable interrupts on receive.
	uart.Bus.IE.Set(kendryte.UARTHS_IE_RXWM)

	intr := interrupt.New(kendryte.IRQ_UARTHS, UART0.handleInterrupt)
	intr.SetPriority(5)
	intr.Enable()
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	rxdata := uart.Bus.RXDATA.Get()
	c := byte(rxdata)
	if uint32(c) != rxdata {
		// The rxdata has other bits set than just the low 8 bits. This probably
		// means that the 'empty' flag is set, which indicates there is no data
		// to be read and the byte is garbage. Ignore this byte.
		return
	}
	uart.Receive(c)
}

func (uart UART) WriteByte(c byte) {
	for uart.Bus.TXDATA.Get()&kendryte.UARTHS_TXDATA_FULL != 0 {
	}

	uart.Bus.TXDATA.Set(uint32(c))
}
