//go:build nrf && !nrf9160
// +build nrf,!nrf9160

package machine

import (
	"device/nrf"
	"runtime/interrupt"
)

// UART on the NRF.
type UART struct {
	*nrf.UART_Type
	Buffer *RingBuffer
	interrupt.Interrupt
}

// UART
var (
	// UART0 is the hardware UART (can be configured as EasyDMA) on the NRF SoC.
	UART0  = &_UART0
	_UART0 = UART{UART_Type: nrf.UART0, Buffer: NewRingBuffer()}
)

func init() {
	UART0.Interrupt = interrupt.New(nrf.IRQ_UART0, _UART0.handleInterrupt)
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
	if uart.UART_Type.EVENTS_RXDRDY.Get() != 0 {
		uart.Receive(byte(uart.UART_Type.RXD.Get()))
		uart.UART_Type.EVENTS_RXDRDY.Set(0x0)
	}
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	uart.SetBaudRate(config.BaudRate)

	// Set TX and RX pins
	if config.TX == 0 && config.RX == 0 {
		// Use default pins
		uart.setPins(UART_TX_PIN, UART_RX_PIN)
	} else {
		uart.setPins(config.TX, config.RX)
	}

	uart.UART_Type.ENABLE.Set(nrf.UART_ENABLE_ENABLE_Enabled)
	uart.UART_Type.TASKS_STARTTX.Set(1)
	uart.UART_Type.TASKS_STARTRX.Set(1)
	uart.UART_Type.INTENSET.Set(nrf.UART_INTENSET_RXDRDY_Msk)

	// Enable RX IRQ.
	uart.Interrupt.SetPriority(0xc0) // low priority
	uart.Interrupt.Enable()
}

// SetBaudRate sets the communication speed for the UART.
func (uart *UART) SetBaudRate(br uint32) {
	// Magic: calculate 'baudrate' register from the input number.
	// Every value listed in the datasheet will be converted to the
	// correct register value, except for 192600. I suspect the value
	// listed in the nrf52 datasheet (0x0EBED000) is incorrectly rounded
	// and should be 0x0EBEE000, as the nrf51 datasheet lists the
	// nonrounded value 0x0EBEDFA4.
	// Some background:
	// https://devzone.nordicsemi.com/f/nordic-q-a/391/uart-baudrate-register-values/2046#2046
	rate := uint32((uint64(br/400)*uint64(400*0xffffffff/16000000) + 0x800) & 0xffffff000)

	uart.UART_Type.BAUDRATE.Set(rate)
}

// WriteByte writes a byte of data to the UART.
func (uart *UART) WriteByte(c byte) error {
	uart.UART_Type.EVENTS_TXDRDY.Set(0)
	uart.UART_Type.TXD.Set(uint32(c))
	for uart.UART_Type.EVENTS_TXDRDY.Get() == 0 {
	}
	return nil
}

// I2C on the NRF.
type I2C struct {
	Bus *nrf.TWI_Type
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = &_I2C0
	I2C1 = &_I2C1
	_I2C0 = I2C{Bus: nrf.TWI0}
	_I2C1 = I2C{Bus: nrf.TWI1}
)

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c *I2C) Tx(addr uint16, w, r []byte) (err error) {

	// Tricky stop condition.
	// After reads, the stop condition is generated implicitly with a shortcut.
	// After writes not followed by reads and in the case of errors, stop must be generated explicitly.

	i2c.Bus.ADDRESS.Set(uint32(addr))

	if len(w) != 0 {
		i2c.Bus.TASKS_STARTTX.Set(1) // start transmission for writing
		for _, b := range w {
			if err = i2c.writeByte(b); err != nil {
				i2c.signalStop()
				return
			}
		}
	}

	if len(r) != 0 {
		// To trigger suspend task when a byte is received
		i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND)
		i2c.Bus.TASKS_STARTRX.Set(1) // re-start transmission for reading
		for i := range r {           // read each char
			if i+1 == len(r) {
				// To trigger stop task when last byte is received, set before resume task.
				i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_STOP)
			}
			if i > 0 {
				i2c.Bus.TASKS_RESUME.Set(1) // re-start transmission for reading
			}
			if r[i], err = i2c.readByte(); err != nil {
				i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND_Disabled)
				i2c.signalStop()
				return
			}
		}
		i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND_Disabled)
	}

	// Stop explicitly when no reads were executed, stoping unconditionally would be a mistake.
	// It may execute after I2C peripheral has already been stopped by the shortcut in the read block,
	// so stop task will trigger first thing in a subsequent transaction, hanging it.
	if len(r) == 0 {
		i2c.signalStop()
	}

	return
}

// writeByte writes a single byte to the I2C bus and waits for confirmation.
func (i2c *I2C) writeByte(data byte) error {
	i2c.Bus.TXD.Set(uint32(data))
	for i2c.Bus.EVENTS_TXDSENT.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return errI2CBusError
		}
	}
	i2c.Bus.EVENTS_TXDSENT.Set(0)
	return nil
}

// readByte reads a single byte from the I2C bus when it is ready.
func (i2c *I2C) readByte() (byte, error) {
	for i2c.Bus.EVENTS_RXDREADY.Get() == 0 {
		if e := i2c.Bus.EVENTS_ERROR.Get(); e != 0 {
			i2c.Bus.EVENTS_ERROR.Set(0)
			return 0, errI2CBusError
		}
	}
	i2c.Bus.EVENTS_RXDREADY.Set(0)
	return byte(i2c.Bus.RXD.Get()), nil
}

var rngStarted = false

// GetRNG returns 32 bits of non-deterministic random data based on internal thermal noise.
// According to Nordic's documentation, the random output is suitable for cryptographic purposes.
func GetRNG() (ret uint32, err error) {
	// There's no apparent way to check the status of the RNG peripheral's task, so simply start it
	// to avoid deadlocking while waiting for output.
	if !rngStarted {
		nrf.RNG.TASKS_START.Set(1)
		nrf.RNG.SetCONFIG_DERCEN(nrf.RNG_CONFIG_DERCEN_Enabled)
		rngStarted = true
	}

	// The RNG returns one byte at a time, so stack up four bytes into a single uint32 for return.
	for i := 0; i < 4; i++ {
		// Wait for data to be ready.
		for nrf.RNG.EVENTS_VALRDY.Get() == 0 {
		}
		// Append random byte to output.
		ret = (ret << 8) ^ nrf.RNG.GetVALUE()
		// Unset the EVENTS_VALRDY register to avoid reading the same random output twice.
		nrf.RNG.EVENTS_VALRDY.Set(0)
	}

	return ret, nil
}

// ReadTemperature reads the silicon die temperature of the chip. The return
// value is in milli-celsius.
func ReadTemperature() int32 {
	nrf.TEMP.TASKS_START.Set(1)
	for nrf.TEMP.EVENTS_DATARDY.Get() == 0 {
	}
	temp := int32(nrf.TEMP.TEMP.Get()) * 250 // the returned value is in units of 0.25°C
	nrf.TEMP.EVENTS_DATARDY.Set(0)
	return temp
}
