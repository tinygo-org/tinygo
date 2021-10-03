package main

import (
	"machine"
	"time"
)

func main() {
	red := machine.Pin(3)
	red.Configure(machine.PinConfig{Mode: machine.PinOutput})
	red.Low()

	green := machine.Pin(4)
	green.Configure(machine.PinConfig{Mode: machine.PinOutput})
	green.Low()

	blue := machine.Pin(5)
	blue.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blue.Low()

	btn := machine.Pin(2)
	btn.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	btn.SetInterrupt(machine.PinInterruptToggle, func(p machine.Pin) {
		v := btn.Get()
		println("button A toggle:", v)
		red.Set(!v)
	})

	btnB := machine.Pin(6)
	btnB.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	btnB.SetInterrupt(machine.PinInterruptToggle, func(p machine.Pin) {
		v := btnB.Get()
		println("button B toggle:", v)
		green.Set(!v)
	})

	for {
		time.Sleep(10 * time.Hour)
	}
}
