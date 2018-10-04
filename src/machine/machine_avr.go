// +build avr

package machine

import (
	"device/avr"
)

type GPIOMode uint8

const (
	GPIO_INPUT = iota
	GPIO_OUTPUT
)

func (p GPIO) Configure(config GPIOConfig) {
	if config.Mode == GPIO_OUTPUT { // set output bit
		if p.Pin < 8 {
			*avr.DDRD |= 1 << p.Pin
		} else {
			*avr.DDRB |= 1 << (p.Pin - 8)
		}
	} else { // configure input: clear output bit
		if p.Pin < 8 {
			*avr.DDRD &^= 1 << p.Pin
		} else {
			*avr.DDRB &^= 1 << (p.Pin - 8)
		}
	}
}

func (p GPIO) Set(value bool) {
	if value { // set bits
		if p.Pin < 8 {
			*avr.PORTD |= 1 << p.Pin
		} else {
			*avr.PORTB |= 1 << (p.Pin - 8)
		}
	} else { // clear bits
		if p.Pin < 8 {
			*avr.PORTD &^= 1 << p.Pin
		} else {
			*avr.PORTB &^= 1 << (p.Pin - 8)
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	if p.Pin < 8 {
		val := *avr.PIND & (1 << p.Pin)
		return (val > 0)
	} else {
		val := *avr.PINB & (1 << (p.Pin - 8))
		return (val > 0)
	}
}

// InitPWM initializes the registers needed for PWM.
func InitPWM() {
	// use waveform generation
	*avr.TCCR0A |= avr.TCCR0A_WGM00

	// set timer 0 prescale factor to 64
	*avr.TCCR0B |= avr.TCCR0B_CS01 | avr.TCCR0B_CS00

	// set timer 1 prescale factor to 64
	*avr.TCCR1B |= avr.TCCR1B_CS11

	// put timer 1 in 8-bit phase correct pwm mode
	*avr.TCCR1A |= avr.TCCR1A_WGM10

	// set timer 2 prescale factor to 64
	*avr.TCCR2B |= avr.TCCR2B_CS22

	// configure timer 2 for phase correct pwm (8-bit)
	*avr.TCCR2A |= avr.TCCR2A_WGM20
}

// Configure configures a PWM pin for output.
func (pwm PWM) Configure() {
	if pwm.Pin < 8 {
		*avr.DDRD |= 1 << pwm.Pin
	} else {
		*avr.DDRB |= 1 << (pwm.Pin - 8)
	}
}

// Set turns on the duty cycle for a PWM pin using the provided value. On the AVR this is normally a
// 8-bit value ranging from 0 to 255.
func (pwm PWM) Set(value uint16) {
	value8 := value >> 8
	switch pwm.Pin {
	case 3:
		// connect pwm to pin on timer 2, channel B
		*avr.TCCR2A |= avr.TCCR2A_COM2B1
		*avr.OCR2B = avr.RegValue(value8) // set pwm duty
	case 5:
		// connect pwm to pin on timer 0, channel B
		*avr.TCCR0A |= avr.TCCR0A_COM0B1
		*avr.OCR0B = avr.RegValue(value8) // set pwm duty
	case 6:
		// connect pwm to pin on timer 0, channel A
		*avr.TCCR0A |= avr.TCCR0A_COM0A1
		*avr.OCR0A = avr.RegValue(value8) // set pwm duty
	case 9:
		// connect pwm to pin on timer 1, channel A
		*avr.TCCR1A |= avr.TCCR1A_COM1A1
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		*avr.OCR1AL = avr.RegValue(value8) // set pwm duty
	case 10:
		// connect pwm to pin on timer 1, channel B
		*avr.TCCR1A |= avr.TCCR1A_COM1B1
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		*avr.OCR1BL = avr.RegValue(value8) // set pwm duty
	case 11:
		// connect pwm to pin on timer 2, channel A
		*avr.TCCR2A |= avr.TCCR2A_COM2A1
		*avr.OCR2A = avr.RegValue(value8) // set pwm duty
	default:
		panic("Invalid PWM pin")
	}
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	// set a2d prescaler so we are inside the desired 50-200 KHz range at 16MHz.
	*avr.ADCSRA |= (avr.ADCSRA_ADPS2 | avr.ADCSRA_ADPS1 | avr.ADCSRA_ADPS0)

	// enable a2d conversions
	*avr.ADCSRA |= avr.ADCSRA_ADEN
}

// Configure configures a ADCPin to be able to be used to read data.
func (a ADC) Configure() {
	return // no pin specific setup on AVR machine.
}

// Get returns the current value of a ADC pin, in the range 0..0xffff. The AVR
// has an ADC of 10 bits precision so the lower 6 bits will be zero.
func (a ADC) Get() uint16 {
	// set the analog reference (high two bits of ADMUX) and select the
	// channel (low 4 bits), masked to only turn on one ADC at a time.
	// set the ADLAR bit (left-adjusted result) to get a value scaled to 16
	// bits. This has the same effect as shifting the return value left by 6
	// bits.
	*avr.ADMUX = avr.RegValue(avr.ADMUX_REFS0 | avr.ADMUX_ADLAR | (a.Pin & 0x07))

	// start the conversion
	*avr.ADCSRA |= avr.ADCSRA_ADSC

	// ADSC is cleared when the conversion finishes
	for ok := true; ok; ok = (*avr.ADCSRA & avr.ADCSRA_ADSC) > 0 {
	}

	low := uint16(*avr.ADCL)
	high := uint16(*avr.ADCH)
	return uint16(low) | uint16(high<<8)
}

// I2C on the Arduino
type I2C struct {
}

// I2C0 is the only I2C interface on the Arduino.
var I2C0 = I2C{}

// TWI_FREQ is the bus speed. Normally either 100 kHz, or 400 kHz for high-speed bus.
const (
	TWI_FREQ_100KHZ = 100000
	TWI_FREQ_400KHZ = 400000
)

// I2CConfig does not do much of anything on Arduino.
type I2CConfig struct {
	Frequency uint32
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}

	// Activate internal pullups for twi.
	*avr.PORTC |= (avr.DIDR0_ADC4D | avr.DIDR0_ADC5D)

	// Initialize twi prescaler and bit rate.
	*avr.TWSR |= (avr.TWSR_TWPS0 | avr.TWSR_TWPS1)

	// twi bit rate formula from atmega128 manual pg. 204:
	// SCL Frequency = CPU Clock Frequency / (16 + (2 * TWBR))
	// NOTE: TWBR should be 10 or higher for master mode.
	// It is 72 for a 16mhz board with 100kHz TWI
	*avr.TWBR = avr.RegValue(((CPU_FREQUENCY / config.Frequency) - 16) / 2)

	// Enable twi module.
	*avr.TWCR = avr.TWCR_TWEN
}

// Start starts an I2C communication session.
func (i2c I2C) Start() {
	// Clear TWI interrupt flag, put start condition on SDA, and enable TWI.
	*avr.TWCR = (avr.TWCR_TWINT | avr.TWCR_TWSTA | avr.TWCR_TWEN)

	// Wait till start condition is transmitted.
	for (*avr.TWCR & avr.TWCR_TWINT) == 0 {
	}
}

// Stop ends an I2C communication session.
func (i2c I2C) Stop() {
	// Send stop condition.
	*avr.TWCR = (avr.TWCR_TWEN | avr.TWCR_TWINT | avr.TWCR_TWSTO)

	// Wait for stop condition to be executed on bus.
	for (*avr.TWCR & avr.TWCR_TWSTO) == 0 {
	}
}

// WriteTo writes a slice of data bytes to a peripheral with a specific address.
func (i2c I2C) WriteTo(address uint8, data []byte) {
	i2c.Start()

	// Write 7-bit shifted peripheral address plus write flag(0)
	i2c.WriteByte(address << 1)

	for _, v := range data {
		i2c.WriteByte(v)
	}

	i2c.Stop()
}

// ReadFrom reads a slice of data bytes from an I2C peripheral with a specific address.
func (i2c I2C) ReadFrom(address uint8, data []byte) {
	i2c.Start()

	// Write 7-bit shifted peripheral address + read flag(1)
	i2c.WriteByte(address<<1 + 1)

	for i, _ := range data {
		data[i] = i2c.ReadByte()
	}

	i2c.Stop()
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) {
	// Write data to register.
	*avr.TWDR = avr.RegValue(data)

	// Clear TWI interrupt flag and enable TWI.
	*avr.TWCR = (avr.TWCR_TWEN | avr.TWCR_TWINT)

	// Wait till data is transmitted.
	for (*avr.TWCR & avr.TWCR_TWINT) == 0 {
	}
}

// ReadByte reads a single byte from the I2C bus.
func (i2c I2C) ReadByte() byte {
	// Clear TWI interrupt flag and enable TWI.
	*avr.TWCR = (avr.TWCR_TWEN | avr.TWCR_TWINT | avr.TWCR_TWEA)

	// Wait till read request is transmitted.
	for (*avr.TWCR & avr.TWCR_TWINT) == 0 {
	}

	return byte(*avr.TWDR)
}

// UART
var (
	// UART0 is the hardware serial port on the AVR.
	UART0 = &UART{}
)

// Configure the UART on the AVR. Defaults to 9600 baud on Arduino.
func (uart UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 9600
	}

	// Set baud rate based on prescale formula from
	// https://www.microchip.com/webdoc/AVRLibcReferenceManual/FAQ_1faq_wrong_baud_rate.html
	// ((F_CPU + UART_BAUD_RATE * 8L) / (UART_BAUD_RATE * 16L) - 1)
	ps := ((CPU_FREQUENCY+config.BaudRate*8)/(config.BaudRate*16) - 1)
	*avr.UBRR0H = avr.RegValue(ps >> 8)
	*avr.UBRR0L = avr.RegValue(ps & 0xff)

	// enable RX, TX and RX interrupt
	*avr.UCSR0B = avr.UCSR0B_RXEN0 | avr.UCSR0B_TXEN0 | avr.UCSR0B_RXCIE0

	// 8-bits data
	*avr.UCSR0C = avr.UCSR0C_UCSZ01 | avr.UCSR0C_UCSZ00
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	// Wait until UART buffer is not busy.
	for (*avr.UCSR0A & avr.UCSR0A_UDRE0) == 0 {
	}
	*avr.UDR0 = avr.RegValue(c) // send char
	return nil
}

//go:interrupt USART_RX_vect
func handleUSART_RX() {
	// Read register to clear it.
	data := *avr.UDR0

	// Ensure no error.
	if (*avr.UCSR0A & (avr.UCSR0A_FE0 | avr.UCSR0A_DOR0 | avr.UCSR0A_UPE0)) == 0 {
		// Put data from UDR register into buffer.
		bufferPut(byte(data))
	}
}
