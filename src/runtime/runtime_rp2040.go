// +build rp2040

package runtime

import (
	"device/arm"

	"machine"
)

// machineTicks is provided by package machine.
func machineTicks() uint64

type timeUnit uint64

// ticks returns the number of ticks (microseconds) elapsed since power up.
func ticks() timeUnit {
	t := machineTicks()
	return timeUnit(t)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func sleepTicks(d timeUnit) {
	if d == 0 {
		return
	}
	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
	}
}

func waitForEvents() {
	arm.Asm("wfe")
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

// machineInit is provided by package machine.
func machineInit()

func init() {
	machineInit()

	machine.Serial.Configure(machine.UARTConfig{})
}

func postinit() {}

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}
