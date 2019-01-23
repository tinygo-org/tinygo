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

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p GPIO) Set(value bool) {
	if value { // set bits
		port, mask := p.PortMaskSet()
		*port = mask
	} else { // clear bits
		port, mask := p.PortMaskClear()
		*port = mask
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: there are no separate pin set/clear registers on the AVR. The
// returned mask is only valid as long as no other pin in the same port has been
// changed.
func (p GPIO) PortMaskSet() (*avr.RegValue, avr.RegValue) {
	port, mask := p.getPortMask()
	return port, *port | avr.RegValue(mask)
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
//
// Warning: there are no separate pin set/clear registers on the AVR. The
// returned mask is only valid as long as no other pin in the same port has been
// changed.
func (p GPIO) PortMaskClear() (*avr.RegValue, avr.RegValue) {
	port, mask := p.getPortMask()
	return port, *port &^ avr.RegValue(mask)
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	// set a2d prescaler so we are inside the desired 50-200 KHz range at 16MHz.
	*avr.ADCSRA |= (avr.ADCSRA_ADPS2 | avr.ADCSRA_ADPS1 | avr.ADCSRA_ADPS0)

	// enable a2d conversions
	*avr.ADCSRA |= avr.ADCSRA_ADEN
}

// Configure configures a ADCPin to be able to be used to read data.
func (a ADC) Configure() {
	return // no pin specific setup on AVR machine.
}

// Get returns the current value of a ADC pin, in the range 0..0xffff. The AVR
// has an ADC of 10 bits precision so the lower 6 bits will be zero.
func (a ADC) Get() uint16 {
	// set the analog reference (high two bits of ADMUX) and select the
	// channel (low 4 bits), masked to only turn on one ADC at a time.
	// set the ADLAR bit (left-adjusted result) to get a value scaled to 16
	// bits. This has the same effect as shifting the return value left by 6
	// bits.
	*avr.ADMUX = avr.RegValue(avr.ADMUX_REFS0 | avr.ADMUX_ADLAR | (a.Pin & 0x07))

	// start the conversion
	*avr.ADCSRA |= avr.ADCSRA_ADSC

	// ADSC is cleared when the conversion finishes
	for ok := true; ok; ok = (*avr.ADCSRA & avr.ADCSRA_ADSC) > 0 {
	}

	low := uint16(*avr.ADCL)
	high := uint16(*avr.ADCH)
	return uint16(low) | uint16(high<<8)
}

// I2C on AVR.
type I2C struct {
}

// I2C0 is the only I2C interface on most AVRs.
var I2C0 = I2C{}

// UART
var (
	// UART0 is the hardware serial port on the AVR.
	UART0 = &UART{Buffer: NewRingBuffer()}
)
