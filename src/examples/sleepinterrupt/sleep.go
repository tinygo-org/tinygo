package main

// This example demonstrates how to use pin change interrupts to
// wake up a dormant microcontroller that is in low power state (dormant).

import (
	"machine"
	"time"
)

func main() {
	led := machine.LED
	// Configure the LED, set to Low (which may be on on some boards).
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()
	println("going into dormant state until rising edge on pin", interruptPin)

	// Set an interrupt on this pin. This sends microcontroller to sleep
	// until a signal is received on the interrupt pin.
	interruptPin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	err := machine.Sleep(interruptPin, machine.PinRising)
	if err != nil {
		println("was unable to sleep:", err.Error())
	} else {
		println("awoken!")
		led.High()
	}

	// Make sure the program won't exit.
	for {
		time.Sleep(time.Hour)
	}
}
