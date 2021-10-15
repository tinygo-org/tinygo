// +build sam,atsamd51,matrixportal_m4

package machine

import (
	"device/sam"
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
