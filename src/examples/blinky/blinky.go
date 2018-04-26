
package main

import (
	"machine"
	"runtime"
)

func main() {
	led := machine.GPIO{17} // LED 1 on the PCA10040
	led.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	for {
		println("LED on")
		led.Set(false)
		runtime.Sleep(runtime.Millisecond * 500)

		println("LED off")
		led.Set(true)
		runtime.Sleep(runtime.Millisecond * 500)
	}
}
