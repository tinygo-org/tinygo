// +build stm32

package machine

// Peripheral abstraction layer for I2C on the stm32 family

import (
	"unsafe"
)

// I2CConfig is used to store config info for I2C.
type I2CConfig struct {
	Frequency uint32
	SCL       Pin
	SDA       Pin
}

// Configure is intended to setup the STM32 I2C interface.
func (i2c I2C) Configure(config I2CConfig) {

	// enable clock for I2C
	enableAltFuncClock(unsafe.Pointer(i2c.Bus))

	// init pins
	if config.SCL == 0 && config.SDA == 0 {
		config.SCL = I2C0_SCL_PIN
		config.SDA = I2C0_SDA_PIN
	}
	i2c.configurePins(config)

	// Get I2C baud rate based on the bus speed it's attached to
	var conf uint32 = i2c.getBaudRate(config)

}
