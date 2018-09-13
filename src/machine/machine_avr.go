// +build avr

package machine

import (
	"device/avr"
)

type GPIOMode uint8

const (
	GPIO_INPUT = iota
	GPIO_OUTPUT
)

// LED on the Arduino
const LED = 13

func (p GPIO) Configure(config GPIOConfig) {
	if config.Mode == GPIO_OUTPUT { // set output bit
		if p.Pin < 8 {
			*avr.DDRD |= 1 << p.Pin
		} else {
			*avr.DDRB |= 1 << (p.Pin - 8)
		}
	} else { // configure input: clear output bit
		if p.Pin < 8 {
			*avr.DDRD &^= 1 << p.Pin
		} else {
			*avr.DDRB &^= 1 << (p.Pin - 8)
		}
	}
}

func (p GPIO) Set(value bool) {
	if value { // set bits
		if p.Pin < 8 {
			*avr.PORTD |= 1 << p.Pin
		} else {
			*avr.PORTB |= 1 << (p.Pin - 8)
		}
	} else { // clear bits
		if p.Pin < 8 {
			*avr.PORTD &^= 1 << p.Pin
		} else {
			*avr.PORTB &^= 1 << (p.Pin - 8)
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	if p.Pin < 8 {
		val := *avr.PIND & (1 << p.Pin)
		return (val > 0)
	} else {
		val := *avr.PINB & (1 << (p.Pin - 8))
		return (val > 0)
	}
}
