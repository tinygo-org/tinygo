// Plays 440 Hz on the left channel and 880 Hz on the right channel on an I2S
// DAC such as the PCM5102.
package main

import (
	"machine"
	"math"
)

const sampleRate = 44100

var sine [256]int16

func main() {
	for i := range sine {
		sine[i] = int16(math.Sin(2*math.Pi*float64(i)/float64(len(sine))) * 8000)
	}

	err := machine.I2S0.Configure(machine.I2SConfig{
		SCK:            pinSCK,
		WS:             pinWS,
		SDO:            pinSDO,
		SDI:            machine.NoPin,
		Mode:           machine.I2SModeSource,
		AudioFrequency: sampleRate,
		Stereo:         true,
	})
	if err != nil {
		println("could not configure I2S:", err.Error())
		return
	}

	// Phase steps in 8.24 fixed point, one table entry per 1<<24.
	left := uint32(440 * uint64(len(sine)) << 24 / sampleRate)
	right := 2 * left
	var phaseL, phaseR uint32
	buf := make([]uint32, 64)
	for {
		for i := range buf {
			l := uint16(sine[(phaseL>>24)%uint32(len(sine))])
			r := uint16(sine[(phaseR>>24)%uint32(len(sine))])
			buf[i] = uint32(l) | uint32(r)<<16
			phaseL += left
			phaseR += right
		}
		machine.I2S0.WriteStereo(buf)
	}
}
