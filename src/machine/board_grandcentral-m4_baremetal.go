// +build grandcentral_m4

package machine

import (
	"device/sam"
)

// I2C on the Grand Central M4
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM3_I2CM,
		SERCOM: 3,
	}
	I2C1 = &I2C{
		Bus:    sam.SERCOM6_I2CM,
		SERCOM: 6,
	}
)

// SPI on the Grand Central M4
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM7_SPIM,
		SERCOM: 7,
	}
	SPI1 = SPI{ // SD card
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
)
