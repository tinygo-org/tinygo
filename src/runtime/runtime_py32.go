//go:build py32

package runtime

import (
	"device/arm"

	"machine"
)

//export Reset_Handler
func main() {
	preinit()

	machine.LED4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ConfigureSystemTimer(8e6)

	run()
	exit(0)
}

const shift = 15

func ConfigureSystemTimer(systemFrequencyHz uint32) {
	arm.SetupSystemTimer(systemFrequencyHz / 1000)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks * 1_000_000)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1_000_000)
}

var tickCounter uint64

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return timeUnit(tickCounter)
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

//export SysTick_Handler
func handleSysTick() {
	tickCounter = tickCounter + 1
	//machine.LED4.Set(!machine.LED4.Get())
	machine.LED4.High()
	for i := 0; i < 100; i++ {
		arm.Asm("nop")
	}
	machine.LED4.Low()
}
