// +build fe310,hifive1b

package machine

import "device/sifive"

// SPI on the HiFive1.
var (
	SPI1 = SPI{
		Bus: sifive.QSPI1,
	}
)

// I2C on the HiFive1 rev B.
var (
	I2C0 = I2C{
		Bus: sifive.I2C0,
	}
)
