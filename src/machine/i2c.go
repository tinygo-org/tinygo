// +build avr nrf sam stm32,!stm32f7x2,!stm32l5x2,!stm32l0 fe310 k210

package machine

import (
	"errors"
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
)

// WriteRegister transmits first the register and then the data to the
// peripheral device.
//
// Many I2C-compatible devices are organized in terms of registers. This method
// is a shortcut to easily write to such registers. Also, it only works for
// devices with 7-bit addresses, which is the vast majority.
func (i2c I2C) WriteRegister(address uint8, register uint8, data []byte) error {
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
func (i2c I2C) ReadRegister(address uint8, register uint8, data []byte) error {
	return i2c.Tx(uint16(address), []byte{register}, data)
}
