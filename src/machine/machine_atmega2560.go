//go:build avr && atmega2560

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART0_RX
const irq_USART1_RX = avr.IRQ_USART1_RX
const irq_USART2_RX = avr.IRQ_USART2_RX
const irq_USART3_RX = avr.IRQ_USART3_RX

const (
	portA Pin = iota * 8
	portB
	portC
	portD
	portE
	portF
	portG
	portH
	portJ
	portK
	portL
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
	PD7 = portD + 7
	PE0 = portE + 0
	PE1 = portE + 1
	PE3 = portE + 3
	PE4 = portE + 4
	PE5 = portE + 5
	PE6 = portE + 6
	PF0 = portF + 0
	PF1 = portF + 1
	PF2 = portF + 2
	PF3 = portF + 3
	PF4 = portF + 4
	PF5 = portF + 5
	PF6 = portF + 6
	PF7 = portF + 7
	PG0 = portG + 0
	PG1 = portG + 1
	PG2 = portG + 2
	PG5 = portG + 5
	PH0 = portH + 0
	PH1 = portH + 1
	PH3 = portH + 3
	PH4 = portH + 4
	PH5 = portH + 5
	PH6 = portH + 6
	PJ0 = portJ + 0
	PJ1 = portJ + 1
	PK0 = portK + 0
	PK1 = portK + 1
	PK2 = portK + 2
	PK3 = portK + 3
	PK4 = portK + 4
	PK5 = portK + 5
	PK6 = portK + 6
	PK7 = portK + 7
	PL0 = portL + 0
	PL1 = portL + 1
	PL2 = portL + 2
	PL3 = portL + 3
	PL4 = portL + 4
	PL5 = portL + 5
	PL6 = portL + 6
	PL7 = portL + 7
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
	case p >= PD0 && p <= PD7:
		return avr.PORTD, 1 << uint8(p-portD)
	case p >= PE0 && p <= PE6:
		return avr.PORTE, 1 << uint8(p-portE)
	case p >= PF0 && p <= PF7:
		return avr.PORTF, 1 << uint8(p-portF)
	case p >= PG0 && p <= PG5:
		return avr.PORTG, 1 << uint8(p-portG)
	case p >= PH0 && p <= PH6:
		return avr.PORTH, 1 << uint8(p-portH)
	case p >= PJ0 && p <= PJ1:
		return avr.PORTJ, 1 << uint8(p-portJ)
	case p >= PK0 && p <= PK7:
		return avr.PORTK, 1 << uint8(p-portK)
	case p >= PL0 && p <= PL7:
		return avr.PORTL, 1 << uint8(p-portL)
	default:
		return avr.PORTA, 255
	}
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spdr: avr.SPDR,
	spsr: avr.SPSR,
	sck:  PB1,
	sdo:  PB2,
	sdi:  PB3,
	cs:   PB0}
