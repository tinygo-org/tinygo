// +build sam,atsamd51,pyportal

package machine

import (
	"device/sam"
)

// SPI on the PyPortal.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
	NINA_SPI = SPI0
)
