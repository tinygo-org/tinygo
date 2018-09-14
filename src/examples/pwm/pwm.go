package main

import (
	"machine"
	"time"
)

// This example assumes that an RGB LED is connected to pins 3, 5 and 6 on an Arduino.
// Change the values below to use different pins.
const (
	redPin   = 3
	greenPin = 5
	bluePin  = 6
)

// cycleColor is just a placeholder until math/rand or some equivalent is working.
func cycleColor(color uint8) uint8 {
	if color < 10 {
		return color + 1
	} else if color < 200 {
		return color + 10
	} else {
		return 0
	}
}

func main() {
	machine.InitPWM()

	red := machine.PWM{redPin}
	red.Configure()

	green := machine.PWM{greenPin}
	green.Configure()

	blue := machine.PWM{bluePin}
	blue.Configure()

	var rc uint8
	var gc uint8 = 20
	var bc uint8 = 30

	for {
		rc = cycleColor(rc)
		gc = cycleColor(gc)
		bc = cycleColor(bc)

		red.Set(uint16(rc) << 8)
		green.Set(uint16(gc) << 8)
		blue.Set(uint16(bc) << 8)

		time.Sleep(time.Millisecond * 500)
	}
}
