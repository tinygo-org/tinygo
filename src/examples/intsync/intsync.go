package main

import (
	"device/arm"
	"machine"
	"runtime/sync"
)

func main() {
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	timerch := make(chan struct{})

	// Toggle the LED on every receive from the timer channel.
	go func() {
		for {
			machine.LED.High()
			<-timerch
			machine.LED.Low()
			<-timerch
		}
	}()

	// Send to the timer channel on every timerCond notification.
	go func() {
		for {
			timerCond.Wait()
			timerch <- struct{}{}
		}
	}()

	// timer fires 10 times per second
	arm.SetupSystemTimer(machine.CPUFrequency() / 10)

	select {}
}

// timerCond is a condition variable used to handle the systick interrupts.
var timerCond sync.IntCond

//go:export SysTick_Handler
func timer_isr() {
	timerCond.Notify()
}
