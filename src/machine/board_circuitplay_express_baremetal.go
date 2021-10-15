// +build sam,atsamd21,circuitplay_express

package machine

import (
	"device/sam"
)

// SPI on the Circuit Playground Express.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM3_SPI,
		SERCOM: 3,
	}
)

// I2S on the Circuit Playground Express.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
