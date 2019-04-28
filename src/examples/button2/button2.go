package main

import (
	"machine"
	"time"
)

// This example assumes that you are using the pca10040 board

func main() {
	led1 := machine.LED1
	led1.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led2 := machine.LED2
	led2.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led3 := machine.LED3
	led3.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led4 := machine.LED4
	led4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	button1 := machine.BUTTON1
	button1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	button2 := machine.BUTTON2
	button2.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	button3 := machine.BUTTON3
	button3.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	button4 := machine.BUTTON4
	button4.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	for {
		led1.Set(button1.Get())
		led2.Set(button2.Get())
		led3.Set(button3.Get())
		led4.Set(button4.Get())

		time.Sleep(time.Millisecond * 10)
	}
}
