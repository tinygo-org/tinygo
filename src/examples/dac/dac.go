// Simplistic example using the DAC on the Circuit Playground Express.
//
// To actually use the DAC for producing complex waveforms or samples requires a DMA
// timer-based playback mechanism which is beyond the scope of this example.
package main

import (
	"machine"
	"time"
)

func main() {
	speaker := machine.A0
	speaker.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.DAC0.Configure(machine.DACConfig{})

	data := []uint16{0xFFFF, 0x8000, 0x4000, 0x2000, 0x1000, 0x0000}

	for {
		for _, val := range data {
			play(val)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func play(val uint16) {
	for i := 0; i < 100; i++ {
		machine.DAC0.Set(val)
		time.Sleep(2 * time.Millisecond)

		machine.DAC0.Set(0)
		time.Sleep(2 * time.Millisecond)
	}
}
