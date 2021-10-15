// +build grandcentral_m4

package machine

import (
	"device/sam"
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
