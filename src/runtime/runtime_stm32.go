//go:build stm32

package runtime

import "device/arm"

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}

func waitForEvents() {
	arm.Asm("wfe")
}
