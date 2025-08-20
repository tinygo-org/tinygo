//go:build !baremetal

package machine

import (
	"crypto/rand"
	"errors"
	"slices"
)

// Dummy machine package that calls out to external functions.

const deviceName = "generic"

var (
	USB = &UART{100}
)

// The Serial port always points to the default UART in a simulated environment.
//
// TODO: perhaps this should be a special serial object that outputs via WASI
// stdout calls.
var Serial = hardwareUART0

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

// Generic PWM/timer peripheral. Properties can be configured depending on the
// hardware.
type timerType struct {
	// Static properties.
	instance    int32
	frequency   uint64
	bits        int
	prescalers  []int
	channelPins [][]Pin

	// Configured 'top' value.
	top uint32
}

// Configure the PWM/timer peripheral.
func (t *timerType) Configure(config PWMConfig) error {
	// Note: for very large period values, this multiplication will overflow.
	top := config.Period * t.frequency / 1e9
	if config.Period == 0 {
		top = 0xffff // default for LEDs
	}

	// The maximum value that can be stored with the given number of bits in
	// this timer.
	maxTop := uint64(1)<<uint64(t.bits) - 1

	// Look for an appropriate prescaler value.
	var prescaler int
	for _, div := range t.prescalers {
		if top/uint64(div) <= maxTop {
			prescaler = div
			top = top / uint64(div)
			break
		}
	}
	if prescaler == 0 {
		return ErrPWMPeriodTooLong
	}

	// Set these values as the configuration.
	t.top = uint32(top)
	pwmConfigure(t.instance, float64(t.frequency)/float64(prescaler), uint32(top))

	return nil
}

// Channel returns a PWM channel for the given pin. Note that one channel may be
// shared between multiple pins, and so will have the same duty cycle. If this
// is not desirable, look for a different PWM/timer peripheral or consider using
// a different pin.
func (t *timerType) Channel(pin Pin) (uint8, error) {
	for ch, pins := range t.channelPins {
		// For nrf52xxx chips specifically we can assign any channel to any pin.
		// We use a similar (identical?) logic to the hardware implementation,
		// and pick the first empty channel.
		if pins == nil {
			t.channelPins[ch] = []Pin{pin}
			pwmChannelConfigure(t.instance, int32(ch), pin)
			return uint8(ch), nil
		}

		// Check whether the pin can be used on this channel.
		for _, p := range pins {
			if p == pin {
				pwmChannelConfigure(t.instance, int32(ch), pin)
				return uint8(ch), nil
			}
		}
	}

	return 0, ErrInvalidOutputPin
}

func (t *timerType) Set(channel uint8, value uint32) {
	pwmChannelSet(t.instance, channel, value)
}

// Top returns the current counter top, for use in duty cycle calculation. It
// will only change with a call to Configure or SetPeriod, otherwise it is
// constant.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to Set
// (see Set documentation for more information).
func (t *timerType) Top() uint32 {
	return t.top
}

//export __tinygo_pwm_configure
func pwmConfigure(instance int32, frequency float64, top uint32)

//export __tinygo_pwm_channel_configure
func pwmChannelConfigure(instance, channel int32, pin Pin)

//export __tinygo_pwm_channel_set
func pwmChannelSet(instance int32, channel uint8, value uint32)

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

func (spi *SPI) Configure(config SPIConfig) error {
	spiConfigure(spi.Bus, config.SCK, config.SDO, config.SDI)
	return nil
}

// Transfer writes/reads a single byte using the SPI interface.
func (spi *SPI) Transfer(w byte) (byte, error) {
	return spiTransfer(spi.Bus, w), nil
}

// Tx handles read/write operation for SPI interface. Since SPI is a synchronous write/read
// interface, there must always be the same number of bytes written as bytes read.
// The Tx method knows about this, and offers a few different ways of calling it.
//
// This form sends the bytes in tx buffer, putting the resulting bytes read into the rx buffer.
// Note that the tx and rx buffers must be the same size:
//
//	spi.Tx(tx, rx)
//
// This form sends the tx buffer, ignoring the result. Useful for sending "commands" that return zeros
// until all the bytes in the command packet have been received:
//
//	spi.Tx(tx, nil)
//
// This form sends zeros, putting the result into the rx buffer. Good for reading a "result packet":
//
//	spi.Tx(nil, rx)
func (spi *SPI) Tx(w, r []byte) error {
	var wptr, rptr *byte
	var wlen, rlen int
	if len(w) != 0 {
		wptr = &w[0]
		wlen = len(w)
	}
	if len(r) != 0 {
		rptr = &r[0]
		rlen = len(r)
	}
	spiTX(spi.Bus, wptr, wlen, rptr, rlen)
	return nil
}

//export __tinygo_spi_configure
func spiConfigure(bus uint8, sck Pin, SDO Pin, SDI Pin)

//export __tinygo_spi_transfer
func spiTransfer(bus uint8, w uint8) uint8

//export __tinygo_spi_tx
func spiTX(bus uint8, wptr *byte, wlen int, rptr *byte, rlen int) uint8

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
	Bus     uint8
	PinsSCL []Pin
	PinsSDA []Pin
}

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the I2C interface.
func (i2c *I2C) Configure(config I2CConfig) error {
	if i2c.PinsSCL != nil {
		matchSCL := slices.Index(i2c.PinsSCL, config.SCL) >= 0
		matchSDA := slices.Index(i2c.PinsSDA, config.SDA) >= 0
		if !matchSCL && !matchSDA {
			return errors.New("i2c: SCL and SDA pins are incorrect for this I2C instance")
		} else if !matchSCL {
			return errors.New("i2c: SCL pin is incorrect for this I2C instance")
		} else if !matchSDA {
			return errors.New("i2c: SDA pin is incorrect for this I2C instance")
		}
	}
	if config.Frequency == 0 {
		config.Frequency = 100 * KHz
	}
	i2cConfigure(i2c.Bus, config.SCL, config.SDA, config.Frequency)
	return nil
}

// SetBaudRate sets the I2C frequency.
func (i2c *I2C) SetBaudRate(br uint32) error {
	i2cSetBaudRate(i2c.Bus, br)
	return nil
}

// Tx does a single I2C transaction at the specified address.
func (i2c *I2C) Tx(addr uint16, w, r []byte) error {
	var wptr, rptr *byte
	var wlen, rlen int
	if len(w) != 0 {
		wptr = &w[0]
		wlen = len(w)
	}
	if len(r) != 0 {
		rptr = &r[0]
		rlen = len(r)
	}
	errCode := i2cTransfer(i2c.Bus, addr, wptr, wlen, rptr, rlen)
	switch errCode {
	case 0:
		return nil
	case 1:
		return errI2CNoDevices
	case 2:
		return errI2CMultipleDevices
	case 3:
		return errI2CWrongAddress
	default:
		return errI2CBusError // unknown error code
	}
}

//export __tinygo_i2c_configure
func i2cConfigure(bus uint8, scl Pin, sda Pin, frequency uint32)

//export __tinygo_i2c_set_baud_rate
func i2cSetBaudRate(bus uint8, br uint32)

//export __tinygo_i2c_transfer
func i2cTransfer(bus uint8, addr uint16, w *byte, wlen int, r *byte, rlen int) int

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

var (
	hardwareUART0 = &UART{0}
	hardwareUART1 = &UART{1}
)

// Some objects used by Atmel SAM D chips (samd21, samd51).
// Defined here (without build tag) for convenience.
var (
	sercomUSART0 = UART{0}
	sercomUSART1 = UART{1}
	sercomUSART2 = UART{2}
	sercomUSART3 = UART{3}
	sercomUSART4 = UART{4}
	sercomUSART5 = UART{5}

	sercomSPIM0 = &SPI{0}
	sercomSPIM1 = &SPI{1}
	sercomSPIM2 = &SPI{2}
	sercomSPIM3 = &SPI{3}
	sercomSPIM4 = &SPI{4}
	sercomSPIM5 = &SPI{5}
	sercomSPIM6 = &SPI{6}
	sercomSPIM7 = &SPI{7}
)

// GetRNG returns 32 bits of random data from the WASI random source.
func GetRNG() (uint32, error) {
	var buf [4]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return uint32(buf[0])<<0 | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24, nil
}
