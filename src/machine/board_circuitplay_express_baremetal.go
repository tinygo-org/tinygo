// +build sam,atsamd21,circuitplay_express

package machine

import (
	"device/sam"
)

// I2S on the Circuit Playground Express.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
