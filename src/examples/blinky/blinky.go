
package main

import (
	"machine"
	"runtime"
)

func main() {
	go led1()
	led2()
}

func led1() {
	led := machine.GPIO{machine.LED}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		println("+")
		led.Low()
		runtime.Sleep(runtime.Millisecond * 1000)

		println("-")
		led.High()
		runtime.Sleep(runtime.Millisecond * 1000)
	}
}

func led2() {
	led := machine.GPIO{machine.LED2}
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		println("  +")
		led.Low()
		runtime.Sleep(runtime.Millisecond * 420)

		println("  -")
		led.High()
		runtime.Sleep(runtime.Millisecond * 420)
	}
}
