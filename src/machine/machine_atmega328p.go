// +build avr,atmega328p

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART_RX

// Configure sets the pin to input or output.
func (p Pin) Configure(config PinConfig) {
	if config.Mode == PinOutput { // set output bit
		switch p / 8 {
		case 0: // port B
			avr.DDRB.SetBits(1 << uint8(p))
		case 1: // port C
			avr.DDRC.SetBits(1 << uint8(p-8))
		case 2: // port D
			avr.DDRD.SetBits(1 << uint8(p-16))
		}
	} else { // configure input: clear output bit
		switch p / 8 {
		case 0: // port B
			avr.DDRB.ClearBits(1 << uint8(p))
		case 1: // port C
			avr.DDRC.ClearBits(1 << uint8(p-8))
		case 2: // port D
			avr.DDRD.ClearBits(1 << uint8(p-16))
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	var val uint8
	switch p / 8 {
	case 0: // port B
		val = avr.PINB.Get() & (1 << uint8(p))
	case 1: // port C
		val = avr.PINC.Get() & (1 << uint8(p-8))
	case 2: // port D
		val = avr.PIND.Get() & (1 << uint8(p-16))
	}
	return val != 0
}

func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch p / 8 {
	case 0: // port B
		return avr.PORTB, 1 << uint8(p)
	case 1:
		return avr.PORTC, 1 << uint8(p-8)
	default:
		return avr.PORTD, 1 << uint8(p-16)
	}
}

// InitPWM initializes the registers needed for PWM.
func InitPWM() {
	// use waveform generation
	avr.TCCR0A.SetBits(avr.TCCR0A_WGM00)

	// set timer 0 prescale factor to 64
	avr.TCCR0B.SetBits(avr.TCCR0B_CS01 | avr.TCCR0B_CS00)

	// set timer 1 prescale factor to 64
	avr.TCCR1B.SetBits(avr.TCCR1B_CS11)

	// put timer 1 in 8-bit phase correct pwm mode
	avr.TCCR1A.SetBits(avr.TCCR1A_WGM10)

	// set timer 2 prescale factor to 64
	avr.TCCR2B.SetBits(avr.TCCR2B_CS22)

	// configure timer 2 for phase correct pwm (8-bit)
	avr.TCCR2A.SetBits(avr.TCCR2A_WGM20)
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() error {
	switch pwm.Pin / 8 {
	case 0: // port B
		avr.DDRB.SetBits(1 << uint8(pwm.Pin))
	case 2: // port D
		avr.DDRD.SetBits(1 << uint8(pwm.Pin-16))
	}
	return nil
}

// Set turns on the duty cycle for a PWM pin using the provided value. On the AVR this is normally a
// 8-bit value ranging from 0 to 255.
func (pwm PWM) Set(value uint16) {
	value8 := uint8(value >> 8)
	switch pwm.Pin {
	case PD3:
		// connect pwm to pin on timer 2, channel B
		avr.TCCR2A.SetBits(avr.TCCR2A_COM2B1)
		avr.OCR2B.Set(value8) // set pwm duty
	case PD5:
		// connect pwm to pin on timer 0, channel B
		avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
		avr.OCR0B.Set(value8) // set pwm duty
	case PD6:
		// connect pwm to pin on timer 0, channel A
		avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
		avr.OCR0A.Set(value8) // set pwm duty
	case PB1:
		// connect pwm to pin on timer 1, channel A
		avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		avr.OCR1AL.Set(value8) // set pwm duty
	case PB2:
		// connect pwm to pin on timer 1, channel B
		avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1)
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		avr.OCR1BL.Set(value8) // set pwm duty
	case PB3:
		// connect pwm to pin on timer 2, channel A
		avr.TCCR2A.SetBits(avr.TCCR2A_COM2A1)
		avr.OCR2A.Set(value8) // set pwm duty
	default:
		panic("Invalid PWM pin")
	}
}
