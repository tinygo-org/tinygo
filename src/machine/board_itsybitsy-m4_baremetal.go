// +build sam,atsamd51,itsybitsy_m4

package machine

import (
	"device/sam"
)

// SPI on the ItsyBitsy M4.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)
