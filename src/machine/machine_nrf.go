// +build nrf

package machine

import (
	"device/nrf"
	"errors"
	"runtime/interrupt"
)

var (
	ErrTxInvalidSliceSize = errors.New("SPI write and read slices must be same size")
)

type PinMode uint8

const (
	PinInput         PinMode = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
	PinInputPullup   PinMode = PinInput | (nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinInputPulldown PinMode = PinOutput | (nrf.GPIO_PIN_CNF_PULL_Pulldown << nrf.GPIO_PIN_CNF_PULL_Pos)
	PinOutput        PinMode = (nrf.GPIO_PIN_CNF_DIR_Output << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Disconnect << nrf.GPIO_PIN_CNF_INPUT_Pos)
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	cfg := config.Mode | nrf.GPIO_PIN_CNF_DRIVE_S0S1 | nrf.GPIO_PIN_CNF_SENSE_Disabled
	port, pin := p.getPortPin()
	port.PIN_CNF[pin].Set(uint32(cfg))
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.OUTSET.Set(1 << pin)
	} else {
		port.OUTCLR.Set(1 << pin)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTSET.Reg, 1 << pin
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return &port.OUTCLR.Reg, 1 << pin
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	port, pin := p.getPortPin()
	return (port.IN.Get()>>pin)&1 != 0
}

// UART on the NRF.
type UART struct {
	Buffer *RingBuffer
}

// UART
var (
	// UART0 is the hardware serial port on the NRF.
	UART0 = UART{Buffer: NewRingBuffer()}
)

// Configure the UART.
func (uart UART) Configure(config UARTConfig) {
	// Default baud rate to 115200.
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}

	uart.SetBaudRate(config.BaudRate)

	// Set TX and RX pins from board.
	uart.setPins(UART_TX_PIN, UART_RX_PIN)

	nrf.UART0.ENABLE.Set(nrf.UART_ENABLE_ENABLE_Enabled)
	nrf.UART0.TASKS_STARTTX.Set(1)
	nrf.UART0.TASKS_STARTRX.Set(1)
	nrf.UART0.INTENSET.Set(nrf.UART_INTENSET_RXDRDY_Msk)

	// Enable RX IRQ.
	intr := interrupt.New(nrf.IRQ_UART0, func(intr interrupt.Interrupt) {
		UART0.handleInterrupt()
	})
	intr.SetPriority(0xc0) // low priority
	intr.Enable()
}

// SetBaudRate sets the communication speed for the UART.
func (uart UART) SetBaudRate(br uint32) {
	// Magic: calculate 'baudrate' register from the input number.
	// Every value listed in the datasheet will be converted to the
	// correct register value, except for 192600. I suspect the value
	// listed in the nrf52 datasheet (0x0EBED000) is incorrectly rounded
	// and should be 0x0EBEE000, as the nrf51 datasheet lists the
	// nonrounded value 0x0EBEDFA4.
	// Some background:
	// https://devzone.nordicsemi.com/f/nordic-q-a/391/uart-baudrate-register-values/2046#2046
	rate := uint32((uint64(br/400)*uint64(400*0xffffffff/16000000) + 0x800) & 0xffffff000)

	nrf.UART0.BAUDRATE.Set(rate)
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	nrf.UART0.EVENTS_TXDRDY.Set(0)
	nrf.UART0.TXD.Set(uint32(c))
	for nrf.UART0.EVENTS_TXDRDY.Get() == 0 {
	}
	return nil
}

func (uart UART) handleInterrupt() {
	if nrf.UART0.EVENTS_RXDRDY.Get() != 0 {
		uart.Receive(byte(nrf.UART0.RXD.Get()))
		nrf.UART0.EVENTS_RXDRDY.Set(0x0)
	}
}

// I2C on the NRF.
type I2C struct {
	Bus *nrf.TWI_Type
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = I2C{Bus: nrf.TWI0}
	I2C1 = I2C{Bus: nrf.TWI1}
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}
	// Default I2C pins if not set.
	if config.SDA == 0 && config.SCL == 0 {
		config.SDA = SDA_PIN
		config.SCL = SCL_PIN
	}

	// do config
	sclPort, sclPin := config.SCL.getPortPin()
	sclPort.PIN_CNF[sclPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	sdaPort, sdaPin := config.SDA.getPortPin()
	sdaPort.PIN_CNF[sdaPin].Set((nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos))

	if config.Frequency == TWI_FREQ_400KHZ {
		i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K400)
	} else {
		i2c.Bus.FREQUENCY.Set(nrf.TWI_FREQUENCY_FREQUENCY_K100)
	}

	i2c.Bus.ENABLE.Set(nrf.TWI_ENABLE_ENABLE_Enabled)
	i2c.setPins(config.SCL, config.SDA)
}

// Tx does a single I2C transaction at the specified address.
// It clocks out the given address, writes the bytes in w, reads back len(r)
// bytes and stores them in r, and generates a stop condition on the bus.
func (i2c I2C) Tx(addr uint16, w, r []byte) error {
	i2c.Bus.ADDRESS.Set(uint32(addr))
	if len(w) != 0 {
		i2c.Bus.TASKS_STARTTX.Set(1) // start transmission for writing
		for _, b := range w {
			i2c.writeByte(b)
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
			i2c.Bus.TASKS_RESUME.Set(1) // re-start transmission for reading
			r[i] = i2c.readByte()
		}
	}
	i2c.signalStop()
	i2c.Bus.SHORTS.Set(nrf.TWI_SHORTS_BB_SUSPEND_Disabled)
	return nil
}

// signalStop sends a stop signal when writing or tells the I2C peripheral that
// it must generate a stop condition after the next character is retrieved when
// reading.
func (i2c I2C) signalStop() {
	i2c.Bus.TASKS_STOP.Set(1)
	for i2c.Bus.EVENTS_STOPPED.Get() == 0 {
	}
	i2c.Bus.EVENTS_STOPPED.Set(0)
}

// writeByte writes a single byte to the I2C bus.
func (i2c I2C) writeByte(data byte) {
	i2c.Bus.TXD.Set(uint32(data))
	for i2c.Bus.EVENTS_TXDSENT.Get() == 0 {
	}
	i2c.Bus.EVENTS_TXDSENT.Set(0)
}

// readByte reads a single byte from the I2C bus.
func (i2c I2C) readByte() byte {
	for i2c.Bus.EVENTS_RXDREADY.Get() == 0 {
	}
	i2c.Bus.EVENTS_RXDREADY.Set(0)
	return byte(i2c.Bus.RXD.Get())
}

// SPI on the NRF.
type SPI struct {
	Bus *nrf.SPI_Type
}

// There are 2 SPI interfaces on the NRF5x.
var (
	SPI0 = SPI{Bus: nrf.SPI0}
	SPI1 = SPI{Bus: nrf.SPI1}
)

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	MOSI      Pin
	MISO      Pin
	LSBFirst  bool
	Mode      uint8
}

// Configure is intended to setup the SPI interface.
func (spi SPI) Configure(config SPIConfig) {
	// Disable bus to configure it
	spi.Bus.ENABLE.Set(nrf.SPI_ENABLE_ENABLE_Disabled)

	// set frequency
	var freq uint32

	switch config.Frequency {
	case 125000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K125
	case 250000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K250
	case 500000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K500
	case 1000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M1
	case 2000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M2
	case 4000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M4
	case 8000000:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_M8
	default:
		freq = nrf.SPI_FREQUENCY_FREQUENCY_K500
	}
	spi.Bus.FREQUENCY.Set(freq)

	var conf uint32

	// set bit transfer order
	if config.LSBFirst {
		conf = (nrf.SPI_CONFIG_ORDER_LsbFirst << nrf.SPI_CONFIG_ORDER_Pos)
	}

	// set mode
	switch config.Mode {
	case 0:
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	case 1:
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf |= (nrf.SPI_CONFIG_CPHA_Trailing << nrf.SPI_CONFIG_CPHA_Pos)
	case 2:
		conf |= (nrf.SPI_CONFIG_CPOL_ActiveLow << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	case 3:
		conf |= (nrf.SPI_CONFIG_CPOL_ActiveLow << nrf.SPI_CONFIG_CPOL_Pos)
		conf |= (nrf.SPI_CONFIG_CPHA_Trailing << nrf.SPI_CONFIG_CPHA_Pos)
	default: // to mode
		conf &^= (nrf.SPI_CONFIG_CPOL_ActiveHigh << nrf.SPI_CONFIG_CPOL_Pos)
		conf &^= (nrf.SPI_CONFIG_CPHA_Leading << nrf.SPI_CONFIG_CPHA_Pos)
	}
	spi.Bus.CONFIG.Set(conf)

	// set pins
	spi.setPins(config.SCK, config.MOSI, config.MISO)

	// Re-enable bus now that it is configured.
	spi.Bus.ENABLE.Set(nrf.SPI_ENABLE_ENABLE_Enabled)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	spi.Bus.TXD.Set(uint32(w))
	for spi.Bus.EVENTS_READY.Get() == 0 {
	}
	r := spi.Bus.RXD.Get()
	spi.Bus.EVENTS_READY.Set(0)

	// TODO: handle SPI errors
	return byte(r), nil
}

// Tx handles read/write operation for SPI interface. Since SPI is a syncronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// The Tx method knows about this, and offers a few different ways of calling it.
//
// This form sends the bytes in tx buffer, putting the resulting bytes read into the rx buffer.
// Note that the tx and rx buffers must be the same size:
//
// 		spi.Tx(tx, rx)
//
// This form sends the tx buffer, ignoring the result. Useful for sending "commands" that return zeros
// until all the bytes in the command packet have been received:
//
// 		spi.Tx(tx, nil)
//
// This form sends zeros, putting the result into the rx buffer. Good for reading a "result packet":
//
// 		spi.Tx(nil, rx)
//
func (spi SPI) Tx(w, r []byte) error {
	var err error

	switch {
	case len(w) == 0:
		// read only, so write zero and read a result.
		for i := range r {
			r[i], err = spi.Transfer(0)
			if err != nil {
				return err
			}
		}
	case len(r) == 0:
		// write only
		spi.Bus.TXD.Set(uint32(w[0]))
		w = w[1:]
		for _, b := range w {
			spi.Bus.TXD.Set(uint32(b))
			for spi.Bus.EVENTS_READY.Get() == 0 {
			}
			_ = spi.Bus.RXD.Get()
			spi.Bus.EVENTS_READY.Set(0)
		}
		for spi.Bus.EVENTS_READY.Get() == 0 {
		}
		_ = spi.Bus.RXD.Get()
		spi.Bus.EVENTS_READY.Set(0)

	default:
		// write/read
		if len(w) != len(r) {
			return ErrTxInvalidSliceSize
		}

		for i, b := range w {
			r[i], err = spi.Transfer(b)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
