// Example using the DAC
//
package main

import (
	"machine"
	"time"
)

func main() {
	enable := machine.PA30
	enable.Configure(machine.PinConfig{Mode: machine.PinOutput})
	enable.Set(true)

	speaker := machine.A0
	speaker.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.DAC0.Configure(machine.DACConfig{})

	data := []uint16{4096, 2058, 1024, 512, 256, 128}

	for {
		for _, val := range data {
			play(val)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func play(val uint16) {
	for i := 0; i < 100; i++ {
		machine.DAC0.WriteData(val)
		time.Sleep(2 * time.Millisecond)

		machine.DAC0.WriteData(0)
		time.Sleep(2 * time.Millisecond)
	}
}
