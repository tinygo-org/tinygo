//go:build py32

package runtime

import (
	"device/arm"
	"machine"
)

//export Reset_Handler
func main() {
	preinit()
	machine.Init()
	run()
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks * 1_000_000)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1_000_000)
}

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return timeUnit(machine.GetTicks())
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}
	start := ticks()
	stop := start + d
	for ticks() < stop {
		arm.Asm("wfe")
	}
}

func putchar(c byte) {

}
