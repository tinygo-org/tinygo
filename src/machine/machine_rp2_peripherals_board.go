//go:build (rp2040 || rp2350) && !boardpackage

package machine

import "device/rp"

// I2C on the RP2040/RP2350.
var (
	I2C0  = &_I2C0
	_I2C0 = I2C{
		Bus: rp.I2C0,
	}
	I2C1  = &_I2C1
	_I2C1 = I2C{
		Bus: rp.I2C1,
	}
)

func defaultI2CPins(config *I2CConfig, bus *rp.I2C0_Type) {
	if config.SCL != 0 || config.SDA != 0 {
		return
	}
	switch bus {
	case rp.I2C0:
		config.SCL = I2C0_SCL_PIN
		config.SDA = I2C0_SDA_PIN
	case rp.I2C1:
		config.SCL = I2C1_SCL_PIN
		config.SDA = I2C1_SDA_PIN
	}
}

// SPI on the RP2040/RP2350.
var (
	SPI0 = &SPI{
		Bus: rp.SPI0,
	}
	SPI1 = &SPI{
		Bus: rp.SPI1,
	}
)

func defaultSPIPins(config *SPIConfig, bus *rp.SPI0_Type) {
	zeroPins := config.SCK == 0 && config.SDO == 0
	if !zeroPins || config.SDI != 0 {
		return
	}
	switch bus {
	case rp.SPI0:
		config.SCK = SPI0_SCK_PIN
		config.SDO = SPI0_SDO_PIN
		config.SDI = SPI0_SDI_PIN
	case rp.SPI1:
		config.SCK = SPI1_SCK_PIN
		config.SDO = SPI1_SDO_PIN
		config.SDI = SPI1_SDI_PIN
	}
}

var (
	PWM0 = NewPWM(0)
	PWM1 = NewPWM(1)
	PWM2 = NewPWM(2)
	PWM3 = NewPWM(3)
	PWM4 = NewPWM(4)
	PWM5 = NewPWM(5)
	PWM6 = NewPWM(6)
	PWM7 = NewPWM(7)
)
