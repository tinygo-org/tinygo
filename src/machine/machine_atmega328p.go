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
		if p < 8 {
			avr.DDRD.SetBits(1 << uint8(p))
		} else if p < 14 {
			avr.DDRB.SetBits(1 << uint8(p-8))
		} else {
			avr.DDRC.SetBits(1 << uint8(p-14))
		}
	} else { // configure input: clear output bit
		if p < 8 {
			avr.DDRD.ClearBits(1 << uint8(p))
		} else if p < 14 {
			avr.DDRB.ClearBits(1 << uint8(p-8))
		} else {
			avr.DDRC.ClearBits(1 << uint8(p-14))
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	if p < 8 {
		val := avr.PIND.Get() & (1 << uint8(p))
		return (val > 0)
	} else if p < 14 {
		val := avr.PINB.Get() & (1 << uint8(p-8))
		return (val > 0)
	} else {
		val := avr.PINC.Get() & (1 << uint8(p-14))
		return (val > 0)
	}
}

func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	if p < 8 {
		return avr.PORTD, 1 << uint8(p)
	} else if p < 14 {
		return avr.PORTB, 1 << uint8(p-8)
	} else {
		return avr.PORTC, 1 << uint8(p-14)
	}
}
