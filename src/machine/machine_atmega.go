// +build avr,atmega

package machine

import (
	"device/avr"
	"runtime/interrupt"
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) error {
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
	// NOTE: TWBR should be 10 or higher for controller mode.
	// It is 72 for a 16mhz board with 100kHz TWI
	avr.TWBR.Set(uint8(((CPUFrequency() / config.Frequency) - 16) / 2))

	// Enable twi module.
	avr.TWCR.Set(avr.TWCR_TWEN)

	return nil
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
	interrupt.New(irq_USART0_RX, func(intr interrupt.Interrupt) {
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
