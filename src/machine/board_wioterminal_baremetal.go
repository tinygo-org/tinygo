// +build sam,atsamd51,wioterminal

package machine

import (
	"device/sam"
)

// I2C on the Wio Terminal
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}

	I2C1 = &I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}
)

// SPI on the Wio Terminal
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM5_SPIM,
		SERCOM: 5,
	}

	// RTL8720D
	SPI1 = SPI{
		Bus:    sam.SERCOM0_SPIM,
		SERCOM: 0,
	}

	// SD
	SPI2 = SPI{
		Bus:    sam.SERCOM6_SPIM,
		SERCOM: 6,
	}

	// LCD
	SPI3 = SPI{
		Bus:    sam.SERCOM7_SPIM,
		SERCOM: 7,
	}
)
