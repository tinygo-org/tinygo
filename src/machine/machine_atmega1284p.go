// +build avr,atmega1284p

package machine

import "device/avr"

const irq_USART0_RX = avr.IRQ_USART0_RX

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 20000000
}
