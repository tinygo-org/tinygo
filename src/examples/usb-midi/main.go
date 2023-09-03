package main

import (
	"machine"
	"machine/usb/adc/midi"

	"time"
)

// Try it easily by opening the following site in Chrome.
// https://www.onlinemusictools.com/kb/

const (
	cable    = 0
	channel  = 1
	velocity = 0x40
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	m := midi.Port()
	m.SetRxHandler(func(b []byte) {
		// blink when we receive a MIDI message
		led.Set(!led.Get())
	})

	m.SetTxHandler(func() {
		// blink when we send a MIDI message
		led.Set(!led.Get())
	})

	prev := true
	chords := []struct {
		name  string
		notes []midi.Note
	}{
		{name: "C ", notes: []midi.Note{midi.C4, midi.E4, midi.G4}},
		{name: "G ", notes: []midi.Note{midi.G3, midi.B3, midi.D4}},
		{name: "Am", notes: []midi.Note{midi.A3, midi.C4, midi.E4}},
		{name: "F ", notes: []midi.Note{midi.F3, midi.A3, midi.C4}},
	}
	index := 0

	for {
		current := button.Get()
		if prev != current {
			if current {
				for _, note := range chords[index].notes {
					m.NoteOff(cable, channel, note, velocity)
				}
				index = (index + 1) % len(chords)
			} else {
				for _, note := range chords[index].notes {
					m.NoteOn(cable, channel, note, velocity)
				}
			}
			prev = current
		}
		time.Sleep(100 * time.Millisecond)
	}
}
