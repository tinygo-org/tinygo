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
	nrf.P0.PIN_CNF[p.Pin] = nrf.RegValue(cfg)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	if high {
		nrf.P0.OUTSET = 1 << p.Pin
	} else {
		nrf.P0.OUTCLR = 1 << p.Pin
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	return (nrf.P0.IN>>p.Pin)&1 != 0
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
	nrf.UART0.PSELTXD = UART_TX_PIN
	nrf.UART0.PSELRXD = UART_RX_PIN

	nrf.UART0.ENABLE = nrf.UART_ENABLE_ENABLE_Enabled
	nrf.UART0.TASKS_STARTTX = 1
	nrf.UART0.TASKS_STARTRX = 1
	nrf.UART0.INTENSET = nrf.UART_INTENSET_RXDRDY_Msk

	// Enable RX IRQ.
	arm.EnableIRQ(nrf.IRQ_UARTE0_UART0)
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

//go:export UARTE0_UART0_IRQHandler
func handleUART0() {
	if nrf.UART0.EVENTS_RXDRDY != 0 {
		bufferPut(byte(nrf.UART0.RXD))
		nrf.UART0.EVENTS_RXDRDY = 0x0
	}
}

// There are 2 I2C interfaces on the NRF.
var (
	I2C0 = I2C{Bus: 0}
	I2C1 = I2C{Bus: 1}
)

// register returns the I2C register that matches the current I2C bus.
func (i2c I2C) register() *nrf.TWI_Type {
	if i2c.Bus == 1 {
		return nrf.TWI1
	}
	return nrf.TWI0
}

// Configure is intended to setup the I2C interface.
func (i2c I2C) Configure(config I2CConfig) {
	// Default I2C bus speed is 100 kHz.
	if config.Frequency == 0 {
		config.Frequency = TWI_FREQ_100KHZ
	}
	// Default I2C pins.
	if config.SDA == 0 {
		config.SDA = SDA_PIN
	}
	if config.SCL == 0 {
		config.SCL = SCL_PIN
	}

	// do config
	nrf.P0.PIN_CNF[config.SCL] = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos)

	nrf.P0.PIN_CNF[config.SDA] = (nrf.GPIO_PIN_CNF_DIR_Input << nrf.GPIO_PIN_CNF_DIR_Pos) |
		(nrf.GPIO_PIN_CNF_INPUT_Connect << nrf.GPIO_PIN_CNF_INPUT_Pos) |
		(nrf.GPIO_PIN_CNF_PULL_Pullup << nrf.GPIO_PIN_CNF_PULL_Pos) |
		(nrf.GPIO_PIN_CNF_DRIVE_S0D1 << nrf.GPIO_PIN_CNF_DRIVE_Pos) |
		(nrf.GPIO_PIN_CNF_SENSE_Disabled << nrf.GPIO_PIN_CNF_SENSE_Pos)

	if config.Frequency == TWI_FREQ_400KHZ {
		i2c.register().FREQUENCY = nrf.TWI_FREQUENCY_FREQUENCY_K400
	} else {
		i2c.register().FREQUENCY = nrf.TWI_FREQUENCY_FREQUENCY_K100
	}

	i2c.register().ENABLE = nrf.TWI_ENABLE_ENABLE_Enabled
	i2c.register().PSELSCL = nrf.RegValue(config.SCL)
	i2c.register().PSELSDA = nrf.RegValue(config.SDA)
}

// Start starts an I2C communication session.
// Nothing to do here because start is handled differently for RX and TX on NRF.
// Strictly here for interface compatibility with AVR.
func (i2c I2C) Start() {
}

// Stop ends an I2C communication session.
func (i2c I2C) Stop() {
	i2c.register().TASKS_STOP = 1
	for i2c.register().EVENTS_STOPPED == 0 {
	}
	i2c.register().EVENTS_STOPPED = 0
}

// WriteTo writes a slice of data bytes to a peripheral with a specific address.
func (i2c I2C) WriteTo(address uint8, data []byte) {
	i2c.register().ADDRESS = nrf.RegValue(address)
	i2c.register().TASKS_STARTTX = 1
	for _, v := range data {
		i2c.WriteByte(v)
	}

	// Assume stop after write.
	i2c.Stop()
}

// ReadFrom reads a slice of data bytes from an I2C peripheral with a specific address.
func (i2c I2C) ReadFrom(address uint8, data []byte) {
	i2c.register().ADDRESS = nrf.RegValue(address)
	i2c.register().TASKS_STARTRX = 1
	for i, _ := range data {
		data[i] = i2c.ReadByte()
	}

	// Assume stop after read.
	i2c.Stop()
}

// WriteByte writes a single byte to the I2C bus.
func (i2c I2C) WriteByte(data byte) {
	i2c.register().TXD = nrf.RegValue(data)
	for i2c.register().EVENTS_TXDSENT == 0 {
	}
	i2c.register().EVENTS_TXDSENT = 0
}

// ReadByte reads a single byte from the I2C bus.
func (i2c I2C) ReadByte() byte {
	for i2c.register().EVENTS_RXDREADY == 0 {
	}
	i2c.register().EVENTS_RXDREADY = 0
	return byte(i2c.register().RXD)
}
