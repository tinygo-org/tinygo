package main

import (
	"machine"
	"time"
)

// This example assumes that an analog sensor such as a rotary dial is connected to pin ADC0.
// When the dial is turned past the midway point, the built-in LED will light up.

func main() {
	machine.InitADC()

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	sensor := machine.ADC{Pin: machine.ADC2}
	sensor.Configure(machine.ADCConfig{})

	for {
		val := sensor.Get()
		if val < 0x8000 {
			led.Low()
		} else {
			led.High()
		}
		time.Sleep(time.Millisecond * 100)
	}
}
