package main

import (
	"machine"
	"time"
)

const (
	led    = machine.LED
	button = machine.BUTTON
)

func main() {
	led.ConfigureAsOutput()
	button.ConfigureAsInput()

	for {
		if button.Get() {
			led.Low()
		} else {
			led.High()
		}

		time.Sleep(time.Millisecond * 10)
	}
}
