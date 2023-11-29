//go:build avr && atmega328pb

package machine

import (
	"device/avr"
	"runtime/interrupt"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART0_RX
const irq_USART1_RX = avr.IRQ_USART1_RX

var (
	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),

		dataReg:    avr.UDR1,
		baudRegH:   avr.UBRR1H,
		baudRegL:   avr.UBRR1L,
		statusRegA: avr.UCSR1A,
		statusRegB: avr.UCSR1B,
		statusRegC: avr.UCSR1C,
	}
)

func init() {
	// Register the UART interrupt.
	interrupt.New(irq_USART1_RX, _UART1.handleInterrupt)
}

// I2C0 is the only I2C interface on most AVRs.
var I2C0 = &I2C{
	srReg: avr.TWSR0,
	brReg: avr.TWBR0,
	crReg: avr.TWCR0,
	drReg: avr.TWDR0,
	srPS0: avr.TWSR0_TWPS0,
	srPS1: avr.TWSR0_TWPS1,
	crEN:  avr.TWCR0_TWEN,
	crINT: avr.TWCR0_TWINT,
	crSTO: avr.TWCR0_TWSTO,
	crEA:  avr.TWCR0_TWEA,
	crSTA: avr.TWCR0_TWSTA,
}

var I2C1 = &I2C{
	srReg: avr.TWSR1,
	brReg: avr.TWBR1,
	crReg: avr.TWCR1,
	drReg: avr.TWDR1,
	srPS0: avr.TWSR1_TWPS10,
	srPS1: avr.TWSR1_TWPS11,
	crEN:  avr.TWCR1_TWEN1,
	crINT: avr.TWCR1_TWINT1,
	crSTO: avr.TWCR1_TWSTO1,
	crEA:  avr.TWCR1_TWEA1,
	crSTA: avr.TWCR1_TWSTA1,
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR0,
	spdr: avr.SPDR0,
	spsr: avr.SPSR0,

	spcrR0:   avr.SPCR0_SPR0,
	spcrR1:   avr.SPCR0_SPR1,
	spcrCPHA: avr.SPCR0_CPHA,
	spcrCPOL: avr.SPCR0_CPOL,
	spcrDORD: avr.SPCR0_DORD,
	spcrSPE:  avr.SPCR0_SPE,
	spcrMSTR: avr.SPCR0_MSTR,

	spsrI2X:  avr.SPSR0_SPI2X,
	spsrSPIF: avr.SPSR0_SPIF,

	sck: PB5,
	sdo: PB3,
	sdi: PB4,
	cs:  PB2,
}

var SPI1 = SPI{
	spcr: avr.SPCR1,
	spdr: avr.SPDR1,
	spsr: avr.SPSR1,

	spcrR0:   avr.SPCR1_SPR10,
	spcrR1:   avr.SPCR1_SPR11,
	spcrCPHA: avr.SPCR1_CPHA1,
	spcrCPOL: avr.SPCR1_CPOL1,
	spcrDORD: avr.SPCR1_DORD1,
	spcrSPE:  avr.SPCR1_SPE1,
	spcrMSTR: avr.SPCR1_MSTR1,

	spsrI2X:  avr.SPSR1_SPI2X1,
	spsrSPIF: avr.SPSR1_SPIF1,

	sck: PC1,
	sdo: PE3,
	sdi: PC0,
	cs:  PE2,
}

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PB0 && p <= PB7: // port B
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC0 && p <= PC7: // port C
		return avr.PORTC, 1 << uint8(p-portC)
	case p >= PD0 && p <= PD7: // port D
		return avr.PORTD, 1 << uint8(p-portD)
	default: // port E
		return avr.PORTE, 1 << uint8(p-portE)
	}
}
