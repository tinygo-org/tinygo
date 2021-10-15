// +build sam,atsamd51,feather_m4

package machine

import (
	"device/sam"
)

// SPI on the Feather M4.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)
