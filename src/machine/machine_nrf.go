// +build nrf

package machine

import (
	"device/arm"
	"device/nrf"
)

type GPIOMode uint8

const (
	GPIO_INPUT          = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos)
	GPIO_INPUT_PULLUP   = GPIO_INPUT | (nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos)
	GPIO_INPUT_PULLDOWN = GPIO_INPUT | (nrf.GPIO_PIN_CNF_PULL_Pulldown << nrf.GPIO_PIN_CNF_PULL_Pos)
	GPIO_OUTPUT         = (nrf.GPIO_PIN_CNF_DIR_Output << nrf.GPIO_PIN_CNF_DIR_Pos) | (nrf.GPIO_PIN_CNF_INPUT_Disconnect << nrf.GPIO_PIN_CNF_INPUT_Pos)
)

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	cfg := config.Mode | nrf.GPIO_PIN_CNF_DRIVE_S0S1 | nrf.GPIO_PIN_CNF_SENSE_Disabled
	port, pin := p.getPortPin()
	port.PIN_CNF[pin] = nrf.RegValue(cfg)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	port, pin := p.getPortPin()
	if high {
		port.OUTSET = 1 << pin
	} else {
		port.OUTCLR = 1 << pin
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p GPIO) PortMaskSet() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return (*uint32)(&port.OUTSET), 1 << pin
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p GPIO) PortMaskClear() (*uint32, uint32) {
	port, pin := p.getPortPin()
	return (*uint32)(&port.OUTCLR), 1 << pin
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	port, pin := p.getPortPin()
	return (port.IN>>pin)&1 != 0
}

// UART
var (
	// UART0 is the hardware serial port on the NRF.
	UART0 = &UART{}
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

	nrf.UART0.ENABLE = nrf.UART_ENABLE_ENABLE_Enabled
	nrf.UART0.TASKS_STARTTX = 1
	nrf.UART0.TASKS_STARTRX = 1
	nrf.UART0.INTENSET = nrf.UART_INTENSET_RXDRDY_Msk

	// Enable RX IRQ.
	arm.EnableIRQ(nrf.IRQ_UART0)
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

	nrf.UART0.BAUDRATE = nrf.RegValue(rate)
}

// WriteByte writes a byte of data to the UART.
func (uart UART) WriteByte(c byte) error {
	nrf.UART0.EVENTS_TXDRDY = 0
	nrf.UART0.TXD = nrf.RegValue(c)
	for nrf.UART0.EVENTS_TXDRDY == 0 {
	}
	return nil
}

func (uart UART) handleInterrupt() {
	if nrf.UART0.EVENTS_RXDRDY != 0 {
		bufferPut(byte(nrf.UART0.RXD))
		nrf.UART0.EVENTS_RXDRDY = 0x0
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
	SCL       uint8
	SDA       uint8
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
	sclPort, sclPin := GPIO{config.SCL}.getPortPin()
	sclPort.PIN_CNF[sclPin] = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos)

	sdaPort, sdaPin := GPIO{config.SDA}.getPortPin()
	sdaPort.PIN_CNF[sdaPin] = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos)

	if config.Frequency == TWI_FREQ_400KHZ {
		i2c.Bus.FREQUENCY = nrf.TWI_FREQUENCY_FREQUENCY_K400
	} else {
		i2c.Bus.FREQUENCY = nrf.TWI_FREQUENCY_FREQUENCY_K100
	}

	i2c.Bus.ENABLE = nrf.TWI_ENABLE_ENABLE_Enabled
	i2c.setPins(config.SCL, config.SDA)
}

// WriteTo writes a slice of data bytes to a peripheral with a specific address.
func (i2c I2C) WriteTo(address uint8, data []byte) {
	i2c.Bus.ADDRESS = nrf.RegValue(address)
	i2c.Bus.TASKS_STARTTX = 1
	for _, v := range data {
		i2c.Bus.TXD = nrf.RegValue(v)
		for i2c.Bus.EVENTS_TXDSENT == 0 {
		}
		i2c.Bus.EVENTS_TXDSENT = 0
	}

	// Assume stop after write.
	i2c.Bus.TASKS_STOP = 1
	for i2c.Bus.EVENTS_STOPPED == 0 {
	}
	i2c.Bus.EVENTS_STOPPED = 0
}

// ReadFrom reads a slice of data bytes from an I2C peripheral with a specific address.
func (i2c I2C) ReadFrom(address uint8, data []byte) {
	i2c.Bus.ADDRESS = nrf.RegValue(address)
	i2c.Bus.TASKS_STARTRX = 1
	for i, _ := range data {
		for i2c.Bus.EVENTS_RXDREADY == 0 {
		}
		i2c.Bus.EVENTS_RXDREADY = 0
		data[i] = byte(i2c.Bus.RXD)
	}

	// Assume stop after read.
	i2c.Bus.TASKS_STOP = 1
	for i2c.Bus.EVENTS_STOPPED == 0 {
	}
	i2c.Bus.EVENTS_STOPPED = 0
}
