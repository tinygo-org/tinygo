package main

import (
	"machine"
	"time"
)

// This example assumes that you are using the pca10040 board

func setLED(led *machine.GPIO, state bool) {
	if state {
		led.High()
	} else {
		led.Low()
	}
}

func main() {
	led1 := machine.GPIO{machine.LED1}
	led1.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	led2 := machine.GPIO{machine.LED2}
	led2.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	led3 := machine.GPIO{machine.LED3}
	led3.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	led4 := machine.GPIO{machine.LED4}
	led4.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	button1 := machine.GPIO{machine.BUTTON1}
	button1.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT_PULLUP})

	button2 := machine.GPIO{machine.BUTTON2}
	button2.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT_PULLUP})

	button3 := machine.GPIO{machine.BUTTON3}
	button3.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT_PULLUP})

	button4 := machine.GPIO{machine.BUTTON4}
	button4.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT_PULLUP})

	for {
		led1.Set(button1.Get())
		led2.Set(button2.Get())
		led3.Set(button3.Get())
		led4.Set(button4.Get())

		time.Sleep(time.Millisecond * 10)
	}
}
