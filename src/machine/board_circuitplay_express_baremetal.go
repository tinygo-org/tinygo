// +build sam,atsamd21,circuitplay_express

package machine

import (
	"device/sam"
)

// I2C on the Circuit Playground Express.
var (
	// external device
	I2C0 = &I2C{
		Bus:    sam.SERCOM5_I2CM,
		SERCOM: 5,
	}
	// internal device
	I2C1 = &I2C{
		Bus:    sam.SERCOM1_I2CM,
		SERCOM: 1,
	}
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
