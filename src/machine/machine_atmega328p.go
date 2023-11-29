//go:build avr && atmega328p

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART_RX

// I2C0 is the only I2C interface on most AVRs.
var I2C0 = &I2C{
	srReg: avr.TWSR,
	brReg: avr.TWBR,
	crReg: avr.TWCR,
	drReg: avr.TWDR,
	srPS0: avr.TWSR_TWPS0,
	srPS1: avr.TWSR_TWPS1,
	crEN:  avr.TWCR_TWEN,
	crINT: avr.TWCR_TWINT,
	crSTO: avr.TWCR_TWSTO,
	crEA:  avr.TWCR_TWEA,
	crSTA: avr.TWCR_TWSTA,
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spdr: avr.SPDR,
	spsr: avr.SPSR,

	spcrR0:   avr.SPCR_SPR0,
	spcrR1:   avr.SPCR_SPR1,
	spcrCPHA: avr.SPCR_CPHA,
	spcrCPOL: avr.SPCR_CPOL,
	spcrDORD: avr.SPCR_DORD,
	spcrSPE:  avr.SPCR_SPE,
	spcrMSTR: avr.SPCR_MSTR,

	spsrI2X:  avr.SPSR_SPI2X,
	spsrSPIF: avr.SPSR_SPIF,

	sck: PB5,
	sdo: PB3,
	sdi: PB4,
	cs:  PB2,
}

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PB0 && p <= PB7: // port B
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC0 && p <= PC7: // port C
		return avr.PORTC, 1 << uint8(p-portC)
	default: // port D
		return avr.PORTD, 1 << uint8(p-portD)
	}
}
