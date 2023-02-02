package machine

import "errors"

var (
	ErrTimeoutRNG         = errors.New("machine: RNG Timeout")
	ErrClockRNG           = errors.New("machine: RNG Clock Error")
	ErrSeedRNG            = errors.New("machine: RNG Seed Error")
	ErrInvalidInputPin    = errors.New("machine: invalid input pin")
	ErrInvalidOutputPin   = errors.New("machine: invalid output pin")
	ErrInvalidClockPin    = errors.New("machine: invalid clock pin")
	ErrInvalidDataPin     = errors.New("machine: invalid data pin")
	ErrNoPinChangeChannel = errors.New("machine: no channel available for pin interrupt")
)

// Device is the running program's chip name, such as "ATSAMD51J19A" or
// "nrf52840". It is not the same as the CPU name.
//
// The constant is some hardcoded default value if the program does not target a
// particular chip but instead runs in WebAssembly for example.
const Device = deviceName

// Generic constants.
const (
	KHz = 1000
	MHz = 1000_000
	GHz = 1000_000_000
)

// PinMode sets the direction and pull mode of the pin. For example, PinOutput
// sets the pin as an output and PinInputPullup sets the pin as an input with a
// pull-up.
type PinMode uint8

type PinConfig struct {
	Mode PinMode
}

// Pin is a single pin on a chip, which may be connected to other hardware
// devices. It can either be used directly as GPIO pin or it can be used in
// other peripherals like ADC, I2C, etc.
type Pin uint8

// NoPin explicitly indicates "not a pin". Use this pin if you want to leave one
// of the pins in a peripheral unconfigured (if supported by the hardware).
const NoPin = Pin(0xff)

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

type ADC struct {
	Pin Pin
}
