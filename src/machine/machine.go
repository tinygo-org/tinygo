package machine

import (
	"errors"
)

// These errors may be returned by SPI.Configure, I2C.Configure, or
// UART.Configure. They indicate that one of the pins (input/output) are
// incorrect.
var (
	ErrInvalidOutputPin = errors.New("machine: invalid output pin")
	ErrInvalidInputPin  = errors.New("machine: invalid input pin")
)

type PinConfig struct {
	Mode PinMode
}

// Pin is a single pin on a chip, which may be connected to other hardware
// devices. It can either be used directly as GPIO pin or it can be used in
// other peripherals like ADC, I2C, etc.
type Pin int8

// NoPin explicitly indicates "not a pin". Use this pin if you want to leave one
// of the pins in a peripheral unconfigured (if supported by the hardware).
const NoPin = Pin(-1)

// High sets this GPIO pin to high, assuming it has been configured as an output
// pin. It is hardware dependent (and often undefined) what happens if you set a
// pin to high that is not configured as an output pin.
func (p Pin) High() {
	p.Set(true)
}

// Low sets this GPIO pin to low, assuming it has been configured as an output
// pin. It is hardware dependent (and often undefined) what happens if you set a
// pin to low that is not configured as an output pin.
func (p Pin) Low() {
	p.Set(false)
}

type PWM struct {
	Pin Pin
}

type ADC struct {
	Pin Pin
}
