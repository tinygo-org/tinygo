//go:build stm32

package runtime

import (
	"device/arm"
	"device/stm32"
)

//export Reset_Handler
func main() {
	if stm32.FPUPresent {
		arm.DisableFPU()
	}
	preinit()
	run()
	exit(0)
}

func waitForEvents() {
	arm.Asm("wfe")
}
