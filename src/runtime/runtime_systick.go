//go:build cortexm && (ra4m1 || py32 || qemu || (nxp && mk66f18))

package runtime

import (
	"device/arm"
	"runtime/interrupt"
	"runtime/volatile"
)

// Milliseconds since initSysTick. SysTick interrupts once each millisecond.
var sysTickCount volatile.Register64

// ticks converts cycles to microseconds with a multiply and this shift, because
// Cortex-M0+ has no divide instruction. 1000 << sysTickShift fits in 32 bits.
const sysTickShift = 22

// SysTick reload value for a 1 ms period.
func sysTickReload() uint32 {
	return sysTickFrequency()/1000 - 1
}

// Start SysTick with a 1 ms period at the rate from sysTickFrequency.
func initSysTick() {
	arm.SetupSystemTimer(sysTickReload())
}

//export SysTick_Handler
func handleSysTick() {
	sysTickCount.Set(sysTickCount.Get() + 1)
}

// ticks returns the time in microseconds since initSysTick.
func ticks() timeUnit {
	state := interrupt.Disable()
	milliseconds := sysTickCount.Get()
	current := arm.SYST.SYST_CVR.Get()
	if arm.SCB.ICSR.HasBits(arm.SCB_ICSR_PENDSTSET) {
		current = arm.SYST.SYST_CVR.Get()
		milliseconds++
	}
	interrupt.Restore(state)

	// The product stays below 1000 << sysTickShift, so microseconds is at most
	// 999 and ticks does not step back at a millisecond boundary.
	elapsedCycles := sysTickReload() - current
	microseconds := elapsedCycles * ((1000 << sysTickShift) / (sysTickReload() + 1)) >> sysTickShift
	return timeUnit(milliseconds*1000 + uint64(microseconds))
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1000
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 1000)
}
