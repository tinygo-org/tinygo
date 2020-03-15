// +build attiny85

package machine

import (
	"device/avr"
	"runtime/volatile"
)

// Configure sets the pin to input or output.
func (p Pin) Configure(config PinConfig) {
	if config.Mode == PinOutput { // set output bit
		avr.DDRB.SetBits(1 << uint8(p))
	} else { // configure input: clear output bit
		avr.DDRB.ClearBits(1 << uint8(p))
	}
}

func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	return avr.PORTB, 1 << uint8(p)
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	val := avr.PINB.Get() & (1 << uint8(p))
	return (val > 0)
}
