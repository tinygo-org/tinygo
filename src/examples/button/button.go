package main

import (
	"machine"
	"runtime"
)

func main() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	button := machine.GPIO{8}
	button.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})

	led.High()

	for {
		if button.Get() {
			led.Low()
		} else {
			led.High()
		}

		runtime.Sleep(runtime.Millisecond * 100)
	}
}
