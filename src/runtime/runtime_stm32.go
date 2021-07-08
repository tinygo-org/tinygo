// +build stm32

package runtime

import "github.com/sago35/device/arm"

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
