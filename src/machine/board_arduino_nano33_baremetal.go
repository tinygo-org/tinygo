// +build sam,atsamd21,arduino_nano33

package machine

import (
	"device/sam"
)

// SPI on the Arduino Nano 33.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPI,
		SERCOM: 1,
	}
)

// SPI1 is connected to the NINA-W102 chip on the Arduino Nano 33.
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM2_SPI,
		SERCOM: 2,
	}
	NINA_SPI = SPI1
)

// I2S on the Arduino Nano 33.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
