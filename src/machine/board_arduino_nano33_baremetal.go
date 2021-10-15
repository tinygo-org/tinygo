// +build sam,atsamd21,arduino_nano33

package machine

import (
	"device/sam"
)

// I2S on the Arduino Nano 33.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
