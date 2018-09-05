package main

// This is the most minimal blinky example and should run almost everywhere.

import (
	"machine"
	"runtime"
)

func main() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		led.Low()
		runtime.Sleep(runtime.Millisecond * 500)

		led.High()
		runtime.Sleep(runtime.Millisecond * 500)
	}
}
