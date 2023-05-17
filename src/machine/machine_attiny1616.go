//go:build attiny1616

package machine

import (
	"device/avr"
)

const (
	portA Pin = iota * 8
	portB
	portC
)

const (
	PA0 = portA + 0
	PA1 = portA + 1
	PA2 = portA + 2
	PA3 = portA + 3
	PA4 = portA + 4
	PA5 = portA + 5
	PA6 = portA + 6
	PA7 = portA + 7
	PB0 = portB + 0
	PB1 = portB + 1
	PB2 = portB + 2
	PB3 = portB + 3
	PB4 = portB + 4
	PB5 = portB + 5
	PB6 = portB + 6
	PB7 = portB + 7
	PC0 = portC + 0
	PC1 = portC + 1
	PC2 = portC + 2
	PC3 = portC + 3
	PC4 = portC + 4
	PC5 = portC + 5
	PC6 = portC + 6
	PC7 = portC + 7
)

// getPortMask returns the PORT peripheral and mask for the pin.
func (p Pin) getPortMask() (*avr.PORT_Type, uint8) {
	switch {
	case p >= PA0 && p <= PA7: // port A
		return avr.PORTA, 1 << uint8(p-portA)
	case p >= PB0 && p <= PB7: // port B
		return avr.PORTB, 1 << uint8(p-portB)
	default: // port C
		return avr.PORTC, 1 << uint8(p-portC)
	}
}
