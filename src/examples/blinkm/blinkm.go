// Connects to an BlinkM I2C RGB LED.
package main

import (
	"machine"
	"time"
)

func main() {
	machine.I2CInit()

	// Init BlinkM
	machine.I2CWriteTo(0x09, []byte("o"))

	version := []byte{0, 0}
	machine.I2CWriteTo(0x09, []byte("Z"))
	machine.I2CReadFrom(0x09, version)
	println("Firmware version:", string(version[0]), string(version[1]))

	count := 0
	for {
		machine.I2CWriteTo(0x09, []byte("n"))

		switch count {
		case 0:
			machine.I2CWriteTo(0x09, []byte{0x99, 0xff, 0x00})
			count = 1
		case 1:
			machine.I2CWriteTo(0x09, []byte{0x00, 0x99, 0xff})
			count = 2
		case 2:
			machine.I2CWriteTo(0x09, []byte{0xff, 0x99, 0x00})
			count = 0
		}

		time.Sleep(100 * time.Millisecond)
	}
}
