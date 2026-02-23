package main

import (
	"machine"
	"time"
)

// This example assumes that an analog sensor such as a rotary dial is connected to pin ADC0.
// When the dial is turned past the midway point, the built-in LED will light up.

func main() {
	time.Sleep(time.Second * 1)

	println("ADC initialized")
	//led := machine.LED
	//led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	sensor := machine.ADC{machine.GPIO2}
	sensor.Configure(machine.ADCConfig{})

	println("ADC configured")
	high := sensor.Pin.Get()
	println("pin voltage check (3.3V->true, once before ADC):", high)

	val := sensor.Get()
	println(val)

	n := 0

	for {
		val := sensor.Get()
		if val == 0 {
			n++
			println("ADC read failed ", n)
			time.Sleep(time.Second * 1)
			continue
		}

		println(val)
		time.Sleep(time.Microsecond * 3000)
	}
}
