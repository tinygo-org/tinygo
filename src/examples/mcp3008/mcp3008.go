// Connects to an MCP3008 ADC via SPI.
// Datasheet: https://www.microchip.com/wwwproducts/en/en010530
package main

import (
	"errors"
	"machine"
	"time"
)

// cs is the pin used for Chip Select (CS). Change to whatever is in use on your board.
const cs = machine.Pin(3)

var (
	tx          []byte
	rx          []byte
	val, result uint16
)

func main() {
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 4000000,
		Mode:      3})

	tx = make([]byte, 3)
	rx = make([]byte, 3)

	for {
		val, _ = Read(0)
		println(val)
		time.Sleep(50 * time.Millisecond)
	}
}

// Read analog data from channel
func Read(channel int) (uint16, error) {
	if channel < 0 || channel > 7 {
		return 0, errors.New("Invalid channel for read")
	}

	tx[0] = 0x01
	tx[1] = byte(8+channel) << 4
	tx[2] = 0x00

	cs.Low()
	machine.SPI0.Tx(tx, rx)
	result = uint16((rx[1]&0x3))<<8 + uint16(rx[2])
	cs.High()

	return result, nil
}
