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
type fpioaPullMode uint8
type PinChange uint8

const (
	PinInput PinMode = iota
	PinInputPullUp
	PinInputPullDown
	PinOutput
)

const (
	fpioaPullNone fpioaPullMode = iota
	fpioaPullDown
	fpioaPullUp
)

const (
	PinRising PinChange = iota + 1
	PinFalling
	PinToggle
	PinHigh
	PinLow = 8
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

// Callbacks to be called for GPIOHS pins configured with SetInterrupt.
var pinCallbacks [32]func(Pin)

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {

	// Check if the pin is a GPIOHS pin.
	if p < P16 && p > P47 {
		return ErrInvalidDataPin
	}

	gpioPin := uint8(p) - 16

	// Clear all interrupts.
	kendryte.GPIOHS.RISE_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.FALL_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.HIGH_IE.ClearBits(1 << gpioPin)
	kendryte.GPIOHS.LOW_IE.ClearBits(1 << gpioPin)

	// Clear all the pending bits for this pin.
	kendryte.GPIOHS.RISE_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.FALL_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.HIGH_IP.SetBits(1 << gpioPin)
	kendryte.GPIOHS.LOW_IP.SetBits(1 << gpioPin)

	if callback == nil {
		if pinCallbacks[gpioPin] != nil {
			pinCallbacks[gpioPin] = nil
		}
		return nil
	}

	if pinCallbacks[gpioPin] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}

	pinCallbacks[gpioPin] = callback

	// Enable interrupts.
	if change&PinRising != 0 {
		kendryte.GPIOHS.RISE_IE.SetBits(1 << gpioPin)
	}
	if change&PinFalling != 0 {
		kendryte.GPIOHS.FALL_IE.SetBits(1 << gpioPin)
	}
	if change&PinHigh != 0 {
		kendryte.GPIOHS.HIGH_IE.SetBits(1 << gpioPin)
	}
	if change&PinLow != 0 {
		kendryte.GPIOHS.LOW_IE.SetBits(1 << gpioPin)
	}

	handleInterrupt := func(inter interrupt.Interrupt) {

		pin := uint8(inter.GetNumber() - kendryte.IRQ_GPIOHS0)

		if kendryte.GPIOHS.RISE_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.RISE_IE.ClearBits(1 << pin)
			kendryte.GPIOHS.RISE_IP.SetBits(1 << pin)
			kendryte.GPIOHS.RISE_IE.SetBits(1 << pin)
		}

		if kendryte.GPIOHS.FALL_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.FALL_IE.ClearBits(1 << pin)
			kendryte.GPIOHS.FALL_IP.SetBits(1 << pin)
			kendryte.GPIOHS.FALL_IE.SetBits(1 << pin)
		}

		if kendryte.GPIOHS.HIGH_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.HIGH_IE.ClearBits(1 << pin)
			kendryte.GPIOHS.HIGH_IP.SetBits(1 << pin)
			kendryte.GPIOHS.HIGH_IE.SetBits(1 << pin)
		}

		if kendryte.GPIOHS.LOW_IE.HasBits(1 << pin) {
			kendryte.GPIOHS.LOW_IE.ClearBits(1 << pin)
			kendryte.GPIOHS.LOW_IP.SetBits(1 << pin)
			kendryte.GPIOHS.LOW_IE.SetBits(1 << pin)
		}

		pinCallbacks[pin](Pin(pin))
	}

	var ir interrupt.Interrupt

	switch p {
	case P16:
		ir = interrupt.New(kendryte.IRQ_GPIOHS0, handleInterrupt)
	case P17:
		ir = interrupt.New(kendryte.IRQ_GPIOHS1, handleInterrupt)
	case P18:
		ir = interrupt.New(kendryte.IRQ_GPIOHS2, handleInterrupt)
	case P19:
		ir = interrupt.New(kendryte.IRQ_GPIOHS3, handleInterrupt)
	case P20:
		ir = interrupt.New(kendryte.IRQ_GPIOHS4, handleInterrupt)
	case P21:
		ir = interrupt.New(kendryte.IRQ_GPIOHS5, handleInterrupt)
	case P22:
		ir = interrupt.New(kendryte.IRQ_GPIOHS6, handleInterrupt)
	case P23:
		ir = interrupt.New(kendryte.IRQ_GPIOHS7, handleInterrupt)
	case P24:
		ir = interrupt.New(kendryte.IRQ_GPIOHS8, handleInterrupt)
	case P25:
		ir = interrupt.New(kendryte.IRQ_GPIOHS9, handleInterrupt)
	case P26:
		ir = interrupt.New(kendryte.IRQ_GPIOHS10, handleInterrupt)
	case P27:
		ir = interrupt.New(kendryte.IRQ_GPIOHS11, handleInterrupt)
	case P28:
		ir = interrupt.New(kendryte.IRQ_GPIOHS12, handleInterrupt)
	case P29:
		ir = interrupt.New(kendryte.IRQ_GPIOHS13, handleInterrupt)
	case P30:
		ir = interrupt.New(kendryte.IRQ_GPIOHS14, handleInterrupt)
	case P31:
		ir = interrupt.New(kendryte.IRQ_GPIOHS15, handleInterrupt)
	case P32:
		ir = interrupt.New(kendryte.IRQ_GPIOHS16, handleInterrupt)
	case P33:
		ir = interrupt.New(kendryte.IRQ_GPIOHS17, handleInterrupt)
	case P34:
		ir = interrupt.New(kendryte.IRQ_GPIOHS18, handleInterrupt)
	case P35:
		ir = interrupt.New(kendryte.IRQ_GPIOHS19, handleInterrupt)
	case P36:
		ir = interrupt.New(kendryte.IRQ_GPIOHS20, handleInterrupt)
	case P37:
		ir = interrupt.New(kendryte.IRQ_GPIOHS21, handleInterrupt)
	case P38:
		ir = interrupt.New(kendryte.IRQ_GPIOHS22, handleInterrupt)
	case P39:
		ir = interrupt.New(kendryte.IRQ_GPIOHS23, handleInterrupt)
	case P40:
		ir = interrupt.New(kendryte.IRQ_GPIOHS24, handleInterrupt)
	case P41:
		ir = interrupt.New(kendryte.IRQ_GPIOHS25, handleInterrupt)
	case P42:
		ir = interrupt.New(kendryte.IRQ_GPIOHS26, handleInterrupt)
	case P43:
		ir = interrupt.New(kendryte.IRQ_GPIOHS27, handleInterrupt)
	case P44:
		ir = interrupt.New(kendryte.IRQ_GPIOHS28, handleInterrupt)
	case P45:
		ir = interrupt.New(kendryte.IRQ_GPIOHS29, handleInterrupt)
	case P46:
		ir = interrupt.New(kendryte.IRQ_GPIOHS30, handleInterrupt)
	case P47:
		ir = interrupt.New(kendryte.IRQ_GPIOHS31, handleInterrupt)
	}

	ir.SetPriority(5)
	ir.Enable()

	return nil

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
