//go:build py32

package runtime

import (
	"device/arm"
	"machine"

	"runtime/volatile"
)

var tickCounter volatile.Register64

//export Reset_Handler
func main() {
	preinit()

	configureHSI()

	configureSystemTimer()
	machine.InitSerial()

	run()
	exit(0)
}

// configureSystemTimer configures SysTick for 1ms ticks.
func configureSystemTimer() {
	arm.SetupSystemTimer(machine.CPUFrequency() / 1000)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks * 1000_000)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000_000)
}

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return timeUnit(tickCounter.Get())
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}
	start := ticks()
	stop := start + d
	for ticks() < stop {
		waitForEvents()
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for machine.Serial.Buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

//export SysTick_Handler
func handleSysTick() {
	tickCounter.Set(tickCounter.Get() + 1)
}
