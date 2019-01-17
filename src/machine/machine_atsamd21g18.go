// +build sam,atsamd21g18a

package machine

import (
	"device/sam"
)

const CPU_FREQUENCY = 48000000

// Peripheral abstraction layer for the atsamd21g18.

type GPIOMode uint8

const (
	GPIO_INPUT  = 0
	GPIO_OUTPUT = 1
)

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	sam.PORT.DIRSET0 = (1 << p.Pin)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	if high {
		sam.PORT.OUTSET0 = (1 << p.Pin)
	} else {
		sam.PORT.OUTCLR0 = (1 << p.Pin)
	}
}
