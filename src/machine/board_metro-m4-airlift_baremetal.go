// +build sam,atsamd51,metro_m4_airlift

package machine

import (
	"device/sam"
)

// SPI on the Metro M4.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
	NINA_SPI = SPI0
)

// SPI1 on the Metro M4 on pins 11,12,13
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)
