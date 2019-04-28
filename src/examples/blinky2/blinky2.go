package main

// This blinky is a bit more advanced than blink1, with two goroutines running
// at the same time and blinking a different LED. The delay of led2 is slightly
// less than half of led1, which would be hard to do without some sort of
// concurrency.

import (
	"machine"
	"time"
)

func main() {
	go led1()
	led2()
}

func led1() {
	led := machine.LED1
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		println("+")
		led.Low()
		time.Sleep(time.Millisecond * 1000)

		println("-")
		led.High()
		time.Sleep(time.Millisecond * 1000)
	}
}

func led2() {
	led := machine.LED2
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		println("  +")
		led.Low()
		time.Sleep(time.Millisecond * 420)

		println("  -")
		led.High()
		time.Sleep(time.Millisecond * 420)
	}
}
