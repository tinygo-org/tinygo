// Example using the i2s hardware interface on the Adafruit Circuit Playground Express
// to read data from the onboard MEMS microphone.
package main

import (
	"machine"
)

func main() {
	machine.I2S0.Configure(machine.I2SConfig{
		Mode:        machine.I2SModePDM,
		ClockSource: machine.I2SClockSourceExternal,
		Stereo:      true,
	})

	data := make([]uint32, 64)

	for {
		// get the next group of samples
		machine.I2S0.Read(data)

		println("data", data[0], data[1], data[2], data[4], "...")
	}
}
