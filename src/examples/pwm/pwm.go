package main

import (
	"machine"
	"runtime"
)

// This example assumes that an RGB LED is connected to pins 3, 5 and 6 on an Arduino.
// Change the values below to use different pins.
const (
	redPin   = 3
	greenPin = 5
	bluePin  = 6
)

// cycleColor is just a placeholder until math/rand or some equivalent is working.
func cycleColor(color uint16) uint16 {
	if color < 10 {
		return color + 1
	}
	if color > 30 && color < 200 {
		return color + 10
	}
	if color > 230 {
		return 0
	}
	return 0
}

func main() {
	machine.InitPWM()

	red := machine.PWM{redPin}
	red.Configure()

	green := machine.PWM{greenPin}
	green.Configure()

	blue := machine.PWM{bluePin}
	blue.Configure()

	var rc uint16
	var gc uint16 = 20
	var bc uint16 = 30

	for {
		rc = cycleColor(rc)
		gc = cycleColor(gc)
		bc = cycleColor(bc)

		red.Set(rc)
		green.Set(gc)
		blue.Set(bc)

		runtime.Sleep(runtime.Millisecond * 500)
	}
}
