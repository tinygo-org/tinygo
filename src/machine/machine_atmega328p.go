// +build avr,atmega328p

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART_RX

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

const (
	Mode0 uint8 = iota
	Mode1
	Mode2
	Mode3
)

type SPIClock uint8

const (
	SPI_CLOCK_FCK2   SPIClock = 0
	SPI_CLOCK_FCK4            = 1
	SPI_CLOCK_FCK8            = 2
	SPI_CLOCK_FCK16           = 3
	SPI_CLOCK_FCK32           = 4
	SPI_CLOCK_FCK64           = 5
	SPI_CLOCK_FCK128          = 7

	SPI_CLOCK_MASK   = 0x03
	SPI_2XCLOCK_MASK = 0x01
)

type SPIConfig struct {
	IsSlave  bool
	LSB      bool
	MaxSpeed SPIClock
	Mode     uint8
	SDI      Pin
	SDO      Pin
	SCK      Pin
}

type SPI struct {
}

var SPI0 = SPI{}

func (spi SPI) Configure(config SPIConfig) {
	setMode(config.Mode)

	if config.LSB {
		spi.LSB()
	} else {
		spi.MSB()
	}

	// Invert the SPI2X bit
	config.MaxSpeed ^= 0x1
	avr.SPSR.SetBits(uint8(config.MaxSpeed) & SPI_2XCLOCK_MASK)

	if config.IsSlave {
		spi.Slave(config)
	} else {
		spi.Master(config)
	}
}

func (s SPI) LSB() {
	avr.SPCR.SetBits(avr.SPCR_DORD)
}

func (s SPI) MSB() {
	avr.SPCR.ClearBits(avr.SPCR_DORD)
}

func (spi SPI) Master(config SPIConfig) {
	avr.DDRB.SetBits(uint8(config.SDO) | uint8(config.SCK))        // set sdo, sck as output, all other input
	avr.DDRB.ClearBits(1 << 4)                                     // sck is high when idle
	avr.SPCR.SetBits(avr.SPCR_MSTR | avr.SPCR_SPR0 | avr.SPCR_SPE) // set master, set clock rate fck/16, enable spi
}

func (spi SPI) Slave(s SPIConfig) {
	avr.DDRB.SetBits(1 << uint8(s.SDI))            // set sdi output, all other input
	avr.SPCR.ClearBits(avr.SPCR_MSTR)              // set slave
	avr.SPCR.SetBits(avr.SPCR_SPR0 | avr.SPCR_SPE) // set clock rate fck/16, enable spi
}

func (SPI) Transfer(b byte) byte {
	avr.SPDR.Set(uint8(b))

	waitForRegisterShift()

	return byte(avr.SPDR.Reg)
}

func (s SPI) Receive() byte {
	waitForRegisterShift()

	return byte(avr.SPDR.Reg)
}

func (s SPI) Send(b byte) {
	avr.SPDR.Set(uint8(b))

	waitForRegisterShift()
}

func waitForRegisterShift() {
	for !avr.SPSR.HasBits(avr.SPSR_SPIF) {
	}
}

func setMode(mode uint8) {
	switch mode {
	case 0:
		avr.SPCR.ClearBits(avr.SPCR_CPOL)
		avr.SPCR.ClearBits(avr.SPCR_CPHA)
	case 1:
		avr.SPCR.ClearBits(avr.SPCR_CPOL)
		avr.SPCR.SetBits(avr.SPCR_CPHA)
	case 2:
		avr.SPCR.SetBits(avr.SPCR_CPOL)
		avr.SPCR.ClearBits(avr.SPCR_CPHA)
	case 3:
		avr.SPCR.SetBits(avr.SPCR_CPOL)
		avr.SPCR.SetBits(avr.SPCR_CPHA)
	default:
		avr.SPCR.ClearBits(avr.SPCR_CPOL)
		avr.SPCR.ClearBits(avr.SPCR_CPHA)
	}
}
