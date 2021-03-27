// +build stm32

package runtime

import (
	"device/arm"
	"machine"
)

type timeUnit int64

const asyncScheduler = false

func postinit() {}

//export Reset_Handler
func main() {
	preinit()
	run()
	abort()
}

func init() {
	machine.InitializeClocks()
	machine.UART0.Configure(machine.UARTConfig{})
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

// number of ticks (microseconds) since start.
func ticks() timeUnit {
	return timeUnit(machine.Ticks())
}

func sleepTicks(d timeUnit) {
	// If there is a scheduler, we sleep until any kind of CPU event up to
	// a maximum of the requested sleep duration.
	//
	// The scheduler will call again if there is nothing to do and a further
	// sleep is required.
	if hasScheduler {
		machine.TimerSleep(ticksToNanoseconds(d))
		return
	}

	// There's no scheduler, so we sleep until at least the requested number
	// of ticks has passed.
	end := ticks() + d
	for ticks() < end {
		machine.TimerSleep(ticksToNanoseconds(d))
	}
}

// Wait for interrupt or event
func waitForEvents() {
	arm.Asm("wfe")
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * machine.TICKS_PER_NS
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / machine.TICKS_PER_NS)
}
