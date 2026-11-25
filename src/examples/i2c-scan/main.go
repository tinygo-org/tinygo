// Scans the I2C bus once per second and prints the addresses that ACK
// as a grid, like i2cdetect.
package main

import (
	"machine"
	"time"
)

const hexDigits = "0123456789abcdef"

func main() {
	err := machine.I2C0.Configure(machine.I2CConfig{Frequency: 100 * machine.KHz})
	if err != nil {
		println("configure:", err.Error())
	}
	for {
		println("     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f")
		for row := uint16(0); row < 0x80; row += 0x10 {
			var line [4 + 3*16]byte
			line[0] = hexDigits[row>>4]
			line[1] = '0'
			line[2] = ':'
			line[3] = ' '
			for col := uint16(0); col < 0x10; col++ {
				cell := line[4+3*col:]
				addr := row + col
				switch {
				case addr < 0x08 || addr > 0x77:
					// Reserved addresses are not probed.
					cell[0], cell[1] = ' ', ' '
				case machine.I2C0.Tx(addr, nil, nil) == nil:
					cell[0], cell[1] = hexDigits[addr>>4], hexDigits[addr&0xf]
				default:
					cell[0], cell[1] = '-', '-'
				}
				cell[2] = ' '
			}
			println(string(line[:]))
		}
		time.Sleep(time.Second)
	}
}
