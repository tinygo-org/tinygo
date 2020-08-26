package main

import (
	"machine"
	"time"
)

// This example assumes that an RGB LED is connected to pins 3, 5 and 6 on an Arduino.
// Change the values below to use different pins.
const (
	redPin   = machine.D4
	greenPin = machine.D5
	bluePin  = machine.D6
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
	err := red.Configure()
	checkError(err, "failed to configure red pin")

	green := machine.PWM{greenPin}
	err = green.Configure()
	checkError(err, "failed to configure green pin")

	blue := machine.PWM{bluePin}
	err = blue.Configure()
	checkError(err, "failed to configure blue pin")

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

func checkError(err error, msg string) {
	if err != nil {
		print(msg, ": ", err.Error())
		println()
	}
}
