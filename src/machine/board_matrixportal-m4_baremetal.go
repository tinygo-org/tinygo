// +build sam,atsamd51,matrixportal_m4

package machine

import (
	"device/sam"
)

// I2C on the MatrixPortal M4
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM5_I2CM,
		SERCOM: 5,
	}
)

// SPI on the MatrixPortal M4
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM3_SPIM,
		SERCOM: 3, // BUG: SDO on SERCOM1!
	}
	NINA_SPI = SPI0

	SPI1 = SPI{
		Bus:    sam.SERCOM0_SPIM,
		SERCOM: 0,
	}
)
