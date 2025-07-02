package main

// This example demonstrates how to use pin change interrupts.
//
// This is only an example and should not be copied directly in any serious
// circuit, because it only naively implements an important feature: debouncing.
// See: https://en.wikipedia.org/wiki/Switch#Contact_bounce

import (
	"machine"
	"time"
)

const (
	led = machine.LED
)

var lastPress time.Time

func main() {

	// Configure the LED, defaulting to on (usually setting the pin to low will
	// turn the LED on).
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	// Make sure the pin is configured as a pullup to avoid floating inputs.
	// Pullup works for most buttons, as most buttons short to ground when
	// pressed.
	button.Configure(machine.PinConfig{Mode: buttonMode})

	// Set an interrupt on this pin.
	err := button.SetInterrupt(buttonPinChange, func(machine.Pin) {

		// Ignore events that are too close to the last registered press (debouncing)
		if time.Since(lastPress) < 100*time.Millisecond {
			return
		}
		lastPress = time.Now()

		led.Set(!led.Get())
	})
	if err != nil {
		println("could not configure pin interrupt:", err.Error())
	}

	// Make sure the program won't exit.
	for {
		time.Sleep(time.Hour)
	}
}
