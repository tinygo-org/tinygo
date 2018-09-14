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

const (
	COM2A1 = 7
	COM2A0 = 6
	COM2B1 = 5
	COM2B0 = 4
	COM0A1 = 7
	COM0A0 = 6
	COM0B1 = 5
	COM0B0 = 4
	COM1A1 = 7
	COM1A0 = 6
	COM1B1 = 5
	COM1B0 = 4

	CS00 = 0
	CS01 = 1
	CS11 = 1
	CS22 = 2

	WGM10 = 0
	WGM11 = 1
	WGM20 = 0
	WGM21 = 1
)

// InitPWM initializes the registers needed for PWM.
func InitPWM() {
	// use waveform generation
	*avr.TCCR0A |= avr.TCCR0A_WGM0

	// set timer 0 prescale factor to 64
	*avr.TCCR0B |= 1 << CS01
	*avr.TCCR0B |= 1 << CS00

	// set timer 1 prescale factor to 64
	*avr.TCCR1B |= 1 << CS11

	// put timer 1 in 8-bit phase correct pwm mode
	*avr.TCCR1A |= 1 << WGM10

	// set timer 2 prescale factor to 64
	*avr.TCCR2B |= 1 << CS22

	// configure timer 2 for phase correct pwm (8-bit)
	*avr.TCCR2A |= 1 << WGM20
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
	if pwm.Pin < 8 {
		*avr.DDRD |= 1 << pwm.Pin
	} else {
		*avr.DDRB |= 1 << (pwm.Pin - 8)
	}
}

// Set sets the needed register values to turn on the duty cycle for a PWM pin.
func (pwm PWM) Set(value uint8) {
	switch pwm.Pin {
	case 3:
		// connect pwm to pin on timer 2, channel B
		*avr.TCCR2A |= 1 << COM2B1
		*avr.OCR2B = avr.RegValue(value) // set pwm duty
	case 5:
		// connect pwm to pin on timer 0, channel B
		*avr.TCCR0A |= 1 << COM0B1
		*avr.OCR0B = avr.RegValue(value) // set pwm duty
	case 6:
		// connect pwm to pin on timer 0, channel A
		*avr.TCCR0A |= 1 << COM0A1
		*avr.OCR0A = avr.RegValue(value) // set pwm duty
	case 9:
		// connect pwm to pin on timer 1, channel A
		*avr.TCCR1A |= 1 << COM1A1
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		*avr.OCR1AL = avr.RegValue(value) // set pwm duty
	case 10:
		// connect pwm to pin on timer 1, channel B
		*avr.TCCR1A |= 1 << COM1B1
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		*avr.OCR1BL = avr.RegValue(value) // set pwm duty
	case 11:
		// connect pwm to pin on timer 2, channel A
		*avr.TCCR2A |= 1 << COM2A1
		*avr.OCR2A = avr.RegValue(value) // set pwm duty
	default:
		// TODO: handle invalid pin for PWM on Arduino
	}
}
