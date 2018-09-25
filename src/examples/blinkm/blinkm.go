// Connects to an BlinkM I2C RGB LED.
// http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf
package main

import (
	"machine"
	"time"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	// Init BlinkM
	machine.I2C0.WriteTo(0x09, []byte("o"))

	version := []byte{0, 0}
	machine.I2C0.WriteTo(0x09, []byte("Z"))
	machine.I2C0.ReadFrom(0x09, version)
	println("Firmware version:", string(version[0]), string(version[1]))

	count := 0
	for {
		machine.I2C0.WriteTo(0x09, []byte("n"))

		switch count {
		case 0:
			// Crimson
			machine.I2C0.WriteTo(0x09, []byte{0xdc, 0x14, 0x3c})
			count = 1
		case 1:
			// MediumPurple
			machine.I2C0.WriteTo(0x09, []byte{0x93, 0x70, 0xdb})
			count = 2
		case 2:
			// MediumSeaGreen
			machine.I2C0.WriteTo(0x09, []byte{0x3c, 0xb3, 0x71})
			count = 0
		}

		time.Sleep(100 * time.Millisecond)
	}
}
