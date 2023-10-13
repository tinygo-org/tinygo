//go:build atmega || nrf || sam || stm32 || fe310 || k210 || rp2040

package machine

import (
	"errors"
)

// If you are getting a compile error on this line please check to see you've
// correctly implemented the methods on the I2C type. They must match
// the i2cController interface method signatures type to type perfectly.
// If not implementing the I2C type please remove your target from the build tags
// at the top of this file.
var _ interface { // 2
	Configure(config I2CConfig) error
	Tx(addr uint16, w, r []byte) error
	SetBaudRate(br uint32) error
} = (*I2C)(nil)

// TWI_FREQ is the I2C bus speed. Normally either 100 kHz, or 400 kHz for high-speed bus.
//
// Deprecated: use 100 * machine.KHz or 400 * machine.KHz instead.
const (
	TWI_FREQ_100KHZ = 100000
	TWI_FREQ_400KHZ = 400000
)

var (
	errI2CWriteTimeout       = errors.New("I2C timeout during write")
	errI2CReadTimeout        = errors.New("I2C timeout during read")
	errI2CBusReadyTimeout    = errors.New("I2C timeout on bus ready")
	errI2CSignalStartTimeout = errors.New("I2C timeout on signal start")
	errI2CSignalReadTimeout  = errors.New("I2C timeout on signal read")
	errI2CSignalStopTimeout  = errors.New("I2C timeout on signal stop")
	errI2CAckExpected        = errors.New("I2C error: expected ACK not NACK")
	errI2CBusError           = errors.New("I2C bus error")
	errI2COverflow           = errors.New("I2C receive buffer overflow")
	errI2COverread           = errors.New("I2C transmit buffer overflow")
	errI2CNotImplemented     = errors.New("I2C operation not yet implemented")
)

// I2CTargetEvent reflects events on the I2C bus
type I2CTargetEvent uint8

const (
	// I2CReceive indicates target has received a message from the controller.
	I2CReceive I2CTargetEvent = iota

	// I2CRequest indicates the controller is expecting a message from the target.
	I2CRequest

	// I2CFinish indicates the controller has ended the transaction.
	//
	// I2C controllers can chain multiple receive/request messages without
	// relinquishing the bus by doing 'restarts'.  I2CFinish indicates the
	// bus has been relinquished by an I2C 'stop'.
	I2CFinish
)

// I2CMode determines if an I2C peripheral is in Controller or Target mode.
type I2CMode int

const (
	// I2CModeController represents an I2C peripheral in controller mode.
	I2CModeController I2CMode = iota

	// I2CModeTarget represents an I2C peripheral in target mode.
	I2CModeTarget
)

// WriteRegister transmits first the register and then the data to the
// peripheral device.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily write to such registers. Also, it only works for
// devices with 7-bit addresses, which is the vast majority.
func (i2c *I2C) WriteRegister(address uint8, register uint8, data []byte) error {
	buf := make([]uint8, len(data)+1)
	buf[0] = register
	copy(buf[1:], data)
	return i2c.Tx(uint16(address), buf, nil)
}

// ReadRegister transmits the register, restarts the connection as a read
// operation, and reads the response.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily read such registers. Also, it only works for devices
// with 7-bit addresses, which is the vast majority.
func (i2c *I2C) ReadRegister(address uint8, register uint8, data []byte) error {
	return i2c.Tx(uint16(address), []byte{register}, data)
}
