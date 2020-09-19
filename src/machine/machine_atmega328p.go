// +build avr,atmega328p

package machine

import (
	"device/avr"
	"errors"
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

var (
	InvalidClockSpeed = errors.New("In periphal mode only a clock speed of FCK4 or lower is allowed")
)

type SPIClockSpeed uint8

const (
	SPI_CLOCK_FOSC4         SPIClockSpeed = 0
	SPI_CLOCK_FOSC16        SPIClockSpeed = 1
	SPI_CLOCK_FOSC64        SPIClockSpeed = 2
	SPI_CLOCK_FOSC128       SPIClockSpeed = 3
	SPI_CLOCK_FOSC2         SPIClockSpeed = 4
	SPI_CLOCK_FOSC8         SPIClockSpeed = 5
	SPI_CLOCK_FOSC32        SPIClockSpeed = 6
	SPI_CLOCK_FOSC64_Double SPIClockSpeed = 7

	SPI_CLOCK_MASK   uint8 = 0x03
	SPI_2XCLOCK_MASK uint8 = 4

	SPI_CPOL_HIGHIDLE bool = true
)

// SPIConfig
type SPIConfig struct {
	// IsPeriphal set to true, when you are operating not in controller mode
	IsPeriphal bool
	LSB        bool
	ClockSpeed SPIClockSpeed
	Mode       uint8
	SDI        Pin
	SDO        Pin
	SCK        Pin
	CPOL       bool
}

// SPI is for the Serial Peripheral Interface
// Data is taken from http://ww1.microchip.com/downloads/en/DeviceDoc/ATmega48A-PA-88A-PA-168A-PA-328-P-DS-DS40002061A.pdf page 169 and following
type SPI struct {
}

var SPI0 = SPI{}

// Configure uses the given config to setup the SPI interface
func (spi SPI) Configure(config SPIConfig) error {
	// When the SPI is configured as Periphal, the SPI is only ensured to work at fosc/4 or lower. Info is taken from the Datasheet
	if config.IsPeriphal && config.ClockSpeed > SPI_CLOCK_FOSC4 {
		return InvalidClockSpeed
	}

	// Use default pins if not set.
	if config.SCK == 0 && config.SDO == 0 && config.SDI == 0 {
		config.SCK = PB5
		config.SDO = PB3
		config.SDI = PB4
	}

	spi.setMode(config.Mode)

	if config.LSB {
		spi.LSB()
	} else {
		spi.MSB()
	}

	if config.CPOL {
		//When this bit is written to one, SCK is high when idle
		avr.SPCR.SetBits(avr.SPCR_CPOL)
	} else {
		//When this bit is written to zero, SCK is low when idle
		avr.SPCR.ClearBits(avr.SPCR_CPOL)
	}

	// Set the SPI2X: Double SPI Speed bit in Bit 0 of SPSR
	avr.SPSR.SetBits(((uint8(config.ClockSpeed) & SPI_2XCLOCK_MASK) >> 2))

	if config.IsPeriphal {
		spi.Periphal(config)
	} else {
		spi.Controller(config)
	}

	return nil
}

// LSB sets LSB mode
func (SPI) LSB() {
	avr.SPCR.SetBits(avr.SPCR_DORD)
}

// MSB sets MSB mode
func (SPI) MSB() {
	avr.SPCR.ClearBits(avr.SPCR_DORD)
}

// Controller setup the SPI interface as controller
func (SPI) Controller(config SPIConfig) {

	avr.DDRB.SetBits(uint8(config.SDO) | uint8(config.SCK)) // set sdo, sck as output, all other input
	avr.DDRB.ClearBits(1 << 4)                              // sck is high when idle
	// set controller, set clock rate, enable spi
	avr.SPCR.SetBits(avr.SPCR_MSTR | (uint8(config.ClockSpeed) & avr.SPCR_SPR0) | (uint8(config.ClockSpeed) & avr.SPCR_SPR1) | avr.SPCR_SPE)
}

// Periphal setup the SPI interface as periphal
func (SPI) Periphal(config SPIConfig) {
	avr.DDRB.SetBits(uint8(config.SDI)) // set sdi output, all other input
	avr.SPCR.ClearBits(avr.SPCR_MSTR)   // set periphal
	// set clock rate fck/16, enable spi
	avr.SPCR.SetBits((uint8(config.ClockSpeed) & avr.SPCR_SPR0) | (uint8(config.ClockSpeed) & avr.SPCR_SPR1) | avr.SPCR_SPE)
}

// Transfer writes the byte into the register and returns it's content
func (spi SPI) Transfer(b byte) (byte, error) {
	avr.SPDR.Set(uint8(b))

	spi.waitForRegisterShift()

	return byte(avr.SPDR.Reg), nil
}

// Receive reads a byte from the register
func (spi SPI) Receive() byte {
	spi.waitForRegisterShift()

	return byte(avr.SPDR.Reg)
}

// Send writes a byte to the SPDR register
func (spi SPI) Send(b byte) {
	avr.SPDR.Set(uint8(b))

	spi.waitForRegisterShift()
}

// waitForRegisterShift waits until the register has shifted
func (SPI) waitForRegisterShift() {
	for !avr.SPSR.HasBits(avr.SPSR_SPIF) {
	}
}

// setMode sets the DataMode
func (SPI) setMode(mode uint8) {
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
		// default is mode 0
		avr.SPCR.ClearBits(avr.SPCR_CPOL)
		avr.SPCR.ClearBits(avr.SPCR_CPHA)
	}
}
