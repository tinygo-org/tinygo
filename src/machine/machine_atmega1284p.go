// +build avr,atmega1284p

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART0_RX

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 20000000
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
