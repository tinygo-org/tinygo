package main

import (
	"machine"

	"tinygo.org/x/device/arm"
)

var timerCh = make(chan struct{}, 1)

func main() {
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// timer fires 10 times per second
	arm.SetupSystemTimer(machine.CPUFrequency() / 10)

	for {
		machine.LED.Low()
		<-timerCh
		machine.LED.High()
		<-timerCh
	}
}

//export SysTick_Handler
func timer_isr() {
	select {
	case timerCh <- struct{}{}:
	default:
		// The consumer is running behind.
	}
}
