package main

import (
	"machine"
	"runtime"
)

// This example assumes that the button is connected to pin 8. Change the value
// below to use a different pin.
const buttonPin = 8

func main() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	button := machine.GPIO{buttonPin}
	button.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})

	for {
		if button.Get() {
			led.Low()
		} else {
			led.High()
		}

		runtime.Sleep(runtime.Millisecond * 10)
	}
}
