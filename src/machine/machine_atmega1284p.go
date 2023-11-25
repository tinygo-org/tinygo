//go:build avr && atmega1284p

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART0_RX

// Size returns the size of the EEPROM for this machine.
func (e EEPROM) Size() int64 {
	return 4096
}

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 20000000
}

const (
	portA Pin = iota * 8
	portB
	portC
	portD
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
	PD0 = portD + 0
	PD1 = portD + 1
	PD2 = portD + 2
	PD3 = portD + 3
	PD4 = portD + 4
	PD5 = portD + 5
	PD6 = portD + 6
	PD7 = portD + 7
)

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PA0 && p <= PA7:
		return avr.PORTA, 1 << uint8(p-portA)
	case p >= PB0 && p <= PB7:
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC0 && p <= PC7:
		return avr.PORTC, 1 << uint8(p-portC)
	default:
		return avr.PORTD, 1 << uint8(p-portD)
	}
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spsr: avr.SPSR,
	spdr: avr.SPDR,
	sck:  PB7,
	sdo:  PB5,
	sdi:  PB6,
	cs:   PB4}
