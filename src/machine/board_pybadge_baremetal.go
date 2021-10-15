// +build sam,atsamd51,pybadge

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

// SPI on the PyBadge.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)

// TFT SPI on the PyBadge.
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM4_SPIM,
		SERCOM: 4,
	}
)
