package main

import (
	"machine"
	"time"
)

// This example assumes that an analog sensor is connected to pin ADC0.

func main() {
	machine.InitADC()

	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	sensor := machine.ADC{machine.ADC2}
	sensor.Configure()

	for {
		val := sensor.Get()
		if val < 255 {
			led.Low()
		} else {
			led.High()
		}
		println("val is", val)

		time.Sleep(time.Millisecond * 100)
	}
}
