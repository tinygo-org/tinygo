package main

// This example demonstrates how to use a common source file for multi-core
// microcontrollers, and how to safely access a shared peripheral (GPIO) from
// both cores simultaneously.
//
// Note that this example also requires an on-board RGB LED with different GPIO
// pins driving each red, green, and blue color, but you can replace these with
// any GPIO pins if needed. It is also possible for both cores to drive the
// same GPIO pin.

import (
	"machine"
	"time"
)

const (
	ledCore0 = machine.LEDG // green
	ledCore1 = machine.LEDB // blue
)

func initLED() machine.Pin {

	// configure all LEDs as GPIO outputs
	ledConfig := machine.PinConfig{Mode: machine.PinOutput}

	// select which unique pin to use based on our current CPU core
	switch machine.CoreID {

	case machine.Core0:
		ledCore0.Configure(ledConfig)
		return ledCore0

	case machine.Core1:
		ledCore1.Configure(ledConfig)
		return ledCore1

	}
	return machine.NoPin
}

func interval() time.Duration {

	switch machine.CoreID {

	case machine.Core0:
		return 200 * time.Millisecond

	case machine.Core1:
		return 500 * time.Millisecond

	}

	return 0
}

func main() {

	p := initLED()
	d := interval()
	t := time.Now()

	for {
		if time.Since(t) > d {
			t = time.Now()
			p.Set(!p.Get()) // or use Toggle() if target supports it
		}
	}
}
