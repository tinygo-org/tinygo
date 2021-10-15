// +build !baremetal

package machine

// Dummy machine package that calls out to external functions.

var (
	UART0 = &UART{0}
	USB   = &UART{100}
)

// The Serial port always points to the default UART in a simulated environment.
//
// TODO: perhaps this should be a special serial object that outputs via WASI
// stdout calls.
var Serial = UART0

const (
	PinInput PinMode = iota
	PinOutput
	PinInputPullup
	PinInputPulldown
)

func (p Pin) Configure(config PinConfig) {
	gpioConfigure(p, config)
}

func (p Pin) Set(value bool) {
	gpioSet(p, value)
}

func (p Pin) Get() bool {
	return gpioGet(p)
}

//export __tinygo_gpio_configure
func gpioConfigure(pin Pin, config PinConfig)

//export __tinygo_gpio_set
func gpioSet(pin Pin, value bool)

//export __tinygo_gpio_get
func gpioGet(pin Pin) bool

type SPI struct {
	Bus uint8
}

type SPIConfig struct {
	Frequency uint32
	SCK       Pin
	SDO       Pin
	SDI       Pin
	Mode      uint8
}

func (spi SPI) Configure(config SPIConfig) {
	spiConfigure(spi.Bus, config.SCK, config.SDO, config.SDI)
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi SPI) Transfer(w byte) (byte, error) {
	return spiTransfer(spi.Bus, w), nil
}

//export __tinygo_spi_configure
func spiConfigure(bus uint8, sck Pin, SDO Pin, SDI Pin)

//export __tinygo_spi_transfer
func spiTransfer(bus uint8, w uint8) uint8

// InitADC enables support for ADC peripherals.
func InitADC() {
	// Nothing to do here.
}

// Configure configures an ADC pin to be able to be used to read data.
func (adc ADC) Configure(ADCConfig) {
}

// Get reads the current analog value from this ADC peripheral.
func (adc ADC) Get() uint16 {
	return adcRead(adc.Pin)
}

//export __tinygo_adc_read
func adcRead(pin Pin) uint16

// I2C is a generic implementation of the Inter-IC communication protocol.
type I2C struct {
	Bus uint8
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {
	i2cConfigure(i2c.Bus, config.SCL, config.SDA)
	return nil
}

// Tx does a single I2C transaction at the specified address.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	i2cTransfer(i2c.Bus, &w[0], len(w), &r[0], len(r))
	// TODO: do something with the returned error code.
	return nil
}

//export __tinygo_i2c_configure
func i2cConfigure(bus uint8, scl Pin, sda Pin)

//export __tinygo_i2c_transfer
func i2cTransfer(bus uint8, w *byte, wlen int, r *byte, rlen int) int

type UART struct {
	Bus uint8
}

// Configure the UART.
func (uart *UART) Configure(config UARTConfig) {
	uartConfigure(uart.Bus, config.TX, config.RX)
}

// Read from the UART.
func (uart *UART) Read(data []byte) (n int, err error) {
	return uartRead(uart.Bus, &data[0], len(data)), nil
}

// Write to the UART.
func (uart *UART) Write(data []byte) (n int, err error) {
	return uartWrite(uart.Bus, &data[0], len(data)), nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart *UART) Buffered() int {
	return 0
}

// ReadByte reads a single byte from the UART.
func (uart *UART) ReadByte() (byte, error) {
	var b byte
	uartRead(uart.Bus, &b, 1)
	return b, nil
}

// WriteByte writes a single byte to the UART.
func (uart *UART) WriteByte(b byte) error {
	uartWrite(uart.Bus, &b, 1)
	return nil
}

//export __tinygo_uart_configure
func uartConfigure(bus uint8, tx Pin, rx Pin)

//export __tinygo_uart_read
func uartRead(bus uint8, buf *byte, bufLen int) int

//export __tinygo_uart_write
func uartWrite(bus uint8, buf *byte, bufLen int) int

// Some objects used by Atmel SAM D chips (samd21, samd51).
// Defined here (without build tag) for convenience.
var (
	sercomUSART0 = UART{0}
	sercomUSART1 = UART{1}
	sercomUSART2 = UART{2}
	sercomUSART3 = UART{3}
	sercomUSART4 = UART{4}
	sercomUSART5 = UART{5}

	sercomI2CM0 = &I2C{0}
	sercomI2CM1 = &I2C{1}
	sercomI2CM2 = &I2C{2}
	sercomI2CM3 = &I2C{3}
	sercomI2CM4 = &I2C{4}
	sercomI2CM5 = &I2C{5}
	sercomI2CM6 = &I2C{6}
	sercomI2CM7 = &I2C{7}
)
