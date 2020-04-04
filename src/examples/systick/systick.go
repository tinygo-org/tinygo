package main

import (
	"device/arm"
	"machine"
)

func main() {
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// timer fires 10 times per second
	arm.SetupSystemTimer(machine.CPUFrequency() / 10)

	for {
	}
}

var led_state bool

//export SysTick_Handler
func timer_isr() {
	if led_state {
		machine.LED.Low()
	} else {
		machine.LED.High()
	}
	led_state = !led_state
}
