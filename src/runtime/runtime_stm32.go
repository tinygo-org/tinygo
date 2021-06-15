// +build stm32,!stm32h7

package runtime

import "device/arm"

type timeUnit int64

func postinit() {}

//export Reset_Handler
func main() {
	preinit()
	run()
	abort()
}

func waitForEvents() {
	arm.Asm("wfe")
}
