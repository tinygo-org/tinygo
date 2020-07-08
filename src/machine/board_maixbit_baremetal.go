// +build k210,maixbit

package machine

import "device/kendryte"

// SPI on the MAix Bit.
var (
	SPI0 = SPI{
		Bus: kendryte.SPI0,
	}
	SPI1 = SPI{
		Bus: kendryte.SPI1,
	}
)

// I2C on the MAix Bit.
var (
	I2C0 = I2C{
		Bus: kendryte.I2C0,
	}
	I2C1 = I2C{
		Bus: kendryte.I2C1,
	}
	I2C2 = I2C{
		Bus: kendryte.I2C2,
	}
)
