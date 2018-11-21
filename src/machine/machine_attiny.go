// +build avr,attiny

package machine

import (
	"device/avr"
)

// Configure sets the pin to input or output.
func (p GPIO) Configure(config GPIOConfig) {
	if config.Mode == GPIO_OUTPUT { // set output bit
		*avr.DDRB |= 1 << p.Pin
	} else { // configure input: clear output bit
		*avr.DDRB &^= 1 << p.Pin
	}
}

func (p GPIO) getPortMask() (*avr.RegValue, uint8) {
	return avr.PORTB, 1 << p.Pin
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	val := *avr.PINB & (1 << p.Pin)
	return (val > 0)
}

// Configure is a dummy implementation. UART has not been implemented for ATtiny
// devices.
func (uart UART) Configure(config UARTConfig) {
}

// WriteByte is a dummy implementation. UART has not been implemented for ATtiny
// devices.
func (uart UART) WriteByte(c byte) error {
	return nil
}

// Tx is a dummy implementation. I2C has not been implemented for ATtiny
// devices.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	return nil
}
