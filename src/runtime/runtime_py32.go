//go:build py32

package runtime

import (
	"device/py32"
)

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return 0
}

func nanosecondsToTicks(ns int64) timeUnit {
	return 0
}

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return 0
}

func sleepTicks(d timeUnit) {
}

func putchar(c byte) {
	// jenom aby tady něco bylo
	py32.GPIOA.SetMODER_MODE0(py32.GPIO_MODER_MODE0_Output)
}
