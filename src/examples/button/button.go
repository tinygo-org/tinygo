package main

import (
	"machine"
	"time"
)

// This example assumes that the button is connected to pin 8. Change the value
// below to use a different pin.
const (
	led    = machine.LED
	button = machine.Pin(8)
)

func main() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	button.Configure(machine.PinConfig{Mode: machine.PinInput})

	for {
		if button.Get() {
			led.Low()
		} else {
			led.High()
		}

		time.Sleep(time.Millisecond * 10)
	}
}
