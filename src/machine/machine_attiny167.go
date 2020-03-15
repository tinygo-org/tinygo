// +build attiny167

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const (
	PA0 Pin = 0
	PA1 Pin = 1
	PA2 Pin = 2
	PA3 Pin = 3
	PA4 Pin = 4
	PA5 Pin = 5
	PA6 Pin = 6
	PA7 Pin = 7
	PB0 Pin = 8
	PB1 Pin = 9
	PB2 Pin = 10
	PB3 Pin = 11
	PB4 Pin = 12
	PB5 Pin = 13
	PB6 Pin = 14
	PB7 Pin = 15
)

// Configure sets the pin to input or output.
func (p Pin) Configure(config PinConfig) {
	if config.Mode == PinOutput { // set output bit
		if p <= PA7 {
			avr.DDRA.SetBits(1 << uint8(p-PA0))
		} else {
			avr.DDRB.SetBits(1 << uint8(p-PB0))
		}
	} else { // configure input: clear output bit
		if p <= PA7 {
			avr.DDRA.ClearBits(1 << uint8(p-PA0))
		} else {
			avr.DDRB.ClearBits(1 << uint8(p-PB0))
		}
	}
}

func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	if p <= PA7 {
		return avr.PORTA, 1 << uint8(p-PA0)
	} else {
		return avr.PORTB, 1 << uint8(p-PB0)
	}
}
