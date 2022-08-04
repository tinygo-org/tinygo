//go:build attiny85
// +build attiny85

package machine

import (
	"runtime/volatile"

	"tinygo.org/x/device/avr"
)

const (
	PB0 Pin = iota
	PB1
	PB2
	PB3
	PB4
	PB5
)

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	// Very simple for the attiny85, which only has a single port.
	return avr.PORTB, 1 << uint8(p)
}
