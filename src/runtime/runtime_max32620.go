// +build max32620

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

func putchar(c byte) {
	// TODO
	//machine.UART0.WriteByte(c)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func ticks() timeUnit {
	// TODO
	return 0
}

func sleepTicks(d timeUnit) {
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
	}
}

const asyncScheduler = false

func waitForEvents() {
	arm.Asm("wfe")
}
