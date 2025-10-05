//go:build rp2350b

package machine

import "device/rp"

// RP2350B has additional pins.

const (
	GPIO30 Pin = 30 // peripherals: PWM7 channel A, I2C1 SDA
	GPIO31 Pin = 31 // peripherals: PWM7 channel B, I2C1 SCL
	GPIO32 Pin = 32 // peripherals: PWM8 channel A, I2C0 SDA
	GPIO33 Pin = 33 // peripherals: PWM8 channel B, I2C0 SCL
	GPIO34 Pin = 34 // peripherals: PWM9 channel A, I2C1 SDA
	GPIO35 Pin = 35 // peripherals: PWM9 channel B, I2C1 SCL
	GPIO36 Pin = 36 // peripherals: PWM10 channel A, I2C0 SDA
	GPIO37 Pin = 37 // peripherals: PWM10 channel B, I2C0 SCL
	GPIO38 Pin = 38 // peripherals: PWM11 channel A, I2C1 SDA
	GPIO39 Pin = 39 // peripherals: PWM11 channel B, I2C1 SCL
	GPIO40 Pin = 40 // peripherals: PWM8 channel A, I2C0 SDA
	GPIO41 Pin = 41 // peripherals: PWM8 channel B, I2C0 SCL
	GPIO42 Pin = 42 // peripherals: PWM9 channel A, I2C1 SDA
	GPIO43 Pin = 43 // peripherals: PWM9 channel B, I2C1 SCL
	GPIO44 Pin = 44 // peripherals: PWM10 channel A, I2C0 SDA
	GPIO45 Pin = 45 // peripherals: PWM10 channel B, I2C0 SCL
	GPIO46 Pin = 46 // peripherals: PWM11 channel A, I2C1 SDA
	GPIO47 Pin = 47 // peripherals: PWM11 channel B, I2C1 SCL
)

// Analog pins on 2350b.
const (
	ADC0 Pin = GPIO40
	ADC1 Pin = GPIO41
	ADC2 Pin = GPIO42
	ADC3 Pin = GPIO43
	ADC4 Pin = GPIO44
	ADC5 Pin = GPIO45
	ADC6 Pin = GPIO46
	ADC7 Pin = GPIO47
	// Ninth ADC channel.
	thermADC = 48
)

// Additional PWMs on the RP2350B.
var (
	PWM8  = getPWMGroup(8)
	PWM9  = getPWMGroup(9)
	PWM10 = getPWMGroup(10)
	PWM11 = getPWMGroup(11)
)

// validPins confirms that the SPI pin selection is a legitimate one
// for the the 2350b chip.
func (spi *SPI) validPins(config SPIConfig) error {
	var okSDI, okSDO, okSCK bool
	switch spi.Bus {
	case rp.SPI0:
		okSDI = config.SDI == 0 || config.SDI == 4 || config.SDI == 16 || config.SDI == 20 || config.SDI == 32 || config.SDI == 36
		okSDO = config.SDO == 3 || config.SDO == 7 || config.SDO == 19 || config.SDO == 23 || config.SDO == 35 || config.SDO == 39
		okSCK = config.SCK == 2 || config.SCK == 6 || config.SCK == 18 || config.SCK == 22 || config.SCK == 34 || config.SCK == 38
	case rp.SPI1:
		okSDI = config.SDI == 8 || config.SDI == 12 || config.SDI == 24 || config.SDI == 28 || config.SDI == 40 || config.SDI == 44
		okSDO = config.SDO == 11 || config.SDO == 15 || config.SDO == 27 || config.SDO == 31 || config.SDO == 43 || config.SDO == 47
		okSCK = config.SCK == 10 || config.SCK == 14 || config.SCK == 26 || config.SCK == 30 || config.SCK == 42 || config.SCK == 46
	}
	switch {
	case !okSDI:
		return errSPIInvalidSDI
	case !okSDO:
		return errSPIInvalidSDO
	case !okSCK:
		return errSPIInvalidSCK
	}
	return nil
}
