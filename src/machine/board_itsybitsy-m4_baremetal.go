// +build sam,atsamd51,itsybitsy_m4

package machine

import (
	"device/sam"
)

// I2C on the ItsyBitsy M4.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM2_I2CM,
		SERCOM: 2,
	}
)

// SPI on the ItsyBitsy M4.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)
