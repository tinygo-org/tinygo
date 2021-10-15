// +build sam,atsamd51,pybadge

package machine

import (
	"device/sam"
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
