//go:build stm32

package runtime

import "device/arm"

type timeUnit int64

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}

func waitForEvents() {
	arm.Asm("wfe")
}
