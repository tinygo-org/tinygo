//go:build ra4m1

package runtime

import (
	"device/arm"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

const (
	cpuFrequency             = 48_000_000
	cyclesPerMicrosecond     = cpuFrequency / 1_000_000
	microsecondsPerInterrupt = 1_000
	sysTickReload            = cpuFrequency/1_000 - 1
)

var sysTickCount volatile.Register64

// ticks returns the number of microseconds elapsed since power up.
func ticks() timeUnit {
	for {
		state := interrupt.Disable()
		milliseconds := sysTickCount.Get()
		current := arm.SYST.SYST_CVR.Get()
		pending := arm.SCB.ICSR.HasBits(arm.SCB_ICSR_PENDSTSET)
		interrupt.Restore(state)

		if pending {
			continue
		}

		elapsedCycles := uint32(sysTickReload) - current
		microseconds := elapsedCycles / cyclesPerMicrosecond
		return timeUnit(
			milliseconds*microsecondsPerInterrupt + uint64(microseconds),
		)
	}
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}

func sleepTicks(d timeUnit) {
	if d <= 0 {
		return
	}

	if hasScheduler {
		if d >= microsecondsPerInterrupt {
			waitForEvents()
		}
		return
	}

	sleepUntil := ticks() + d
	for ticks() < sleepUntil {
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

func buffered() int {
	return machine.Serial.Buffered()
}

func init() {
	arm.SetupSystemTimer(sysTickReload)
	arm.EnableInterrupts(0)
}

//go:export SysTick_Handler
func handleSysTick() {
	sysTickCount.Set(sysTickCount.Get() + 1)
}

//export Reset_Handler
func main() {
	preinit()
	run()
	exit(0)
}
