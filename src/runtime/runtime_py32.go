//go:build py32

package runtime

import "machine"

//export Reset_Handler
func main() {
	preinit()

	configureHSI()

	initSysTick()
	initStdio()

	run()
	exit(0)
}

func sysTickFrequency() uint32 {
	return machine.CPUFrequency()
}
