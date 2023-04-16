package main

import (
	"machine"
	"machine/usb/midi"

	"encoding/hex"
	"time"
)

// Try it easily by opening the following site in Chrome.
// https://www.onlinemusictools.com/kb/

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	m := midi.Port()
	m.SetHandler(func(b []byte) {
		led.Set(!led.Get())
		println(hex.EncodeToString(b))
		m.Write(b)
	})

	prev := true
	chords := []struct {
		name string
		keys []midi.Note
	}{
		{name: "C ", keys: []midi.Note{midi.C4, midi.E4, midi.G4}},
		{name: "G ", keys: []midi.Note{midi.G3, midi.B3, midi.D4}},
		{name: "Am", keys: []midi.Note{midi.A3, midi.C4, midi.E4}},
		{name: "F ", keys: []midi.Note{midi.F3, midi.A3, midi.C4}},
	}
	index := 0

	for {
		current := button.Get()
		if prev != current {
			led.Set(current)
			if current {
				for _, c := range chords[index].keys {
					m.NoteOff(0, 0, c, 0x40)
				}
				index = (index + 1) % len(chords)
			} else {
				for _, c := range chords[index].keys {
					m.NoteOn(0, 0, c, 0x40)
				}
			}
			prev = current
		}
		time.Sleep(10 * time.Millisecond)
	}
}
