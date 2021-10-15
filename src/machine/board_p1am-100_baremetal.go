// +build sam,atsamd21,p1am_100

package machine

import (
	"device/sam"
)

// I2C on the P1AM-100.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM0_I2CM,
		SERCOM: 0,
	}
)

// SPI on the P1AM-100 is used for Base Controller.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPI,
		SERCOM: 1,
	}
	BASE_CONTROLLER_SPI = SPI0
)

// SPI1 is connected to the SD card slot on the P1AM-100
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM2_SPI,
		SERCOM: 2,
	}
	SDCARD_SPI = SPI1
)
