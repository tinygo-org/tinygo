// +build avr,atmega

package machine

import (
	"device/avr"
	"runtime/interrupt"
	"runtime/volatile"
)

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
func (pwm PWM) Configure() {
	if pwm.Pin < 8 {
		avr.DDRD.SetBits(1 << uint8(pwm.Pin))
	} else {
		avr.DDRB.SetBits(1 << uint8(pwm.Pin-8))
	}
}

// Set turns on the duty cycle for a PWM pin using the provided value. On the AVR this is normally a
// 8-bit value ranging from 0 to 255.
func (pwm PWM) Set(value uint16) {
	value8 := uint8(value >> 8)
	switch pwm.Pin {
	case 3:
		// connect pwm to pin on timer 2, channel B
		avr.TCCR2A.SetBits(avr.TCCR2A_COM2B1)
		avr.OCR2B.Set(value8) // set pwm duty
	case 5:
		// connect pwm to pin on timer 0, channel B
		avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
		avr.OCR0B.Set(value8) // set pwm duty
	case 6:
		// connect pwm to pin on timer 0, channel A
		avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
		avr.OCR0A.Set(value8) // set pwm duty
	case 9:
		// connect pwm to pin on timer 1, channel A
		avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		avr.OCR1AL.Set(value8) // set pwm duty
	case 10:
		// connect pwm to pin on timer 1, channel B
		avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1)
		// this is a 16-bit value, but we only currently allow the low order bits to be set
		avr.OCR1BL.Set(value8) // set pwm duty
	case 11:
		// connect pwm to pin on timer 2, channel A
		avr.TCCR2A.SetBits(avr.TCCR2A_COM2A1)
		avr.OCR2A.Set(value8) // set pwm duty
	default:
		panic("Invalid PWM pin")
	}
}

// I2CConfig is used to store config info for I2C.
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
	avr.PORTC.SetBits((avr.DIDR0_ADC4D | avr.DIDR0_ADC5D))

	// Initialize twi prescaler and bit rate.
	avr.TWSR.SetBits((avr.TWSR_TWPS0 | avr.TWSR_TWPS1))

	// twi bit rate formula from atmega128 manual pg. 204:
	// SCL Frequency = CPU Clock Frequency / (16 + (2 * TWBR))
	// NOTE: TWBR should be 10 or higher for master mode.
	// It is 72 for a 16mhz board with 100kHz TWI
	avr.TWBR.Set(uint8(((CPUFrequency() / config.Frequency) - 16) / 2))

	// Enable twi module.
	avr.TWCR.Set(avr.TWCR_TWEN)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	if len(w) != 0 {
		i2c.start(uint8(addr), true) // start transmission for writing
		for _, b := range w {
			i2c.writeByte(b)
		}
	}
	if len(r) != 0 {
		i2c.start(uint8(addr), false) // re-start transmission for reading
		for i := range r {            // read each char
			r[i] = i2c.readByte()
		}
	}
	if len(w) != 0 || len(r) != 0 {
		// Stop the transmission after it has been started.
		i2c.stop()
	}
	return nil
}

// start starts an I2C communication session.
func (i2c I2C) start(address uint8, write bool) {
	// Clear TWI interrupt flag, put start condition on SDA, and enable TWI.
	avr.TWCR.Set((avr.TWCR_TWINT | avr.TWCR_TWSTA | avr.TWCR_TWEN))

	// Wait till start condition is transmitted.
	for !avr.TWCR.HasBits(avr.TWCR_TWINT) {
	}

	// Write 7-bit shifted peripheral address.
	address <<= 1
	if !write {
		address |= 1 // set read flag
	}
	i2c.writeByte(address)
}

// stop ends an I2C communication session.
func (i2c I2C) stop() {
	// Send stop condition.
	avr.TWCR.Set(avr.TWCR_TWEN | avr.TWCR_TWINT | avr.TWCR_TWSTO)

	// Wait for stop condition to be executed on bus.
	for !avr.TWCR.HasBits(avr.TWCR_TWSTO) {
	}
}

// writeByte writes a single byte to the I2C bus.
func (i2c I2C) writeByte(data byte) {
	// Write data to register.
	avr.TWDR.Set(data)

	// Clear TWI interrupt flag and enable TWI.
	avr.TWCR.Set(avr.TWCR_TWEN | avr.TWCR_TWINT)

	// Wait till data is transmitted.
	for !avr.TWCR.HasBits(avr.TWCR_TWINT) {
	}
}

// readByte reads a single byte from the I2C bus.
func (i2c I2C) readByte() byte {
	// Clear TWI interrupt flag and enable TWI.
	avr.TWCR.Set(avr.TWCR_TWEN | avr.TWCR_TWINT | avr.TWCR_TWEA)

	// Wait till read request is transmitted.
	for !avr.TWCR.HasBits(avr.TWCR_TWINT) {
	}

	return byte(avr.TWDR.Get())
}

// UART on the AVR.
type UART struct {
	Buffer *RingBuffer
}

// Configure the UART on the AVR. Defaults to 9600 baud on Arduino.
func (uart UART) Configure(config UARTConfig) {
	if config.BaudRate == 0 {
		config.BaudRate = 9600
	}

	// Register the UART interrupt.
	interrupt.New(avr.IRQ_USART_RX, func(intr interrupt.Interrupt) {
		// Read register to clear it.
		data := avr.UDR0.Get()

		// Ensure no error.
		if !avr.UCSR0A.HasBits(avr.UCSR0A_FE0 | avr.UCSR0A_DOR0 | avr.UCSR0A_UPE0) {
			// Put data from UDR register into buffer.
			UART0.Receive(byte(data))
		}
	})

	// Set baud rate based on prescale formula from
	// https://www.microchip.com/webdoc/AVRLibcReferenceManual/FAQ_1faq_wrong_baud_rate.html
	// ((F_CPU + UART_BAUD_RATE * 8L) / (UART_BAUD_RATE * 16L) - 1)
	ps := ((CPUFrequency()+config.BaudRate*8)/(config.BaudRate*16) - 1)
	avr.UBRR0H.Set(uint8(ps >> 8))
	avr.UBRR0L.Set(uint8(ps & 0xff))

	// enable RX, TX and RX interrupt
	avr.UCSR0B.Set(avr.UCSR0B_RXEN0 | avr.UCSR0B_TXEN0 | avr.UCSR0B_RXCIE0)

	// 8-bits data
	avr.UCSR0C.Set(avr.UCSR0C_UCSZ01 | avr.UCSR0C_UCSZ00)
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	// Wait until UART buffer is not busy.
	for !avr.UCSR0A.HasBits(avr.UCSR0A_UDRE0) {
	}
	avr.UDR0.Set(c) // send char
	return nil
}
