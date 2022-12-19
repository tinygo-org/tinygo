//go:build rp2040

package machine

import (
	"device/arm"
	"device/rp"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const numTimers = 4

// Alarm0 is reserved for sleeping by tinygo runtime code for RP2040.
// Alarm0 is also IRQ0
const sleepAlarm = 0
const sleepAlarmIRQ = 0

// The minimum sleep duration in μs (ticks)
const minSleep = 10

type timerType struct {
	timeHW   volatile.Register32
	timeLW   volatile.Register32
	timeHR   volatile.Register32
	timeLR   volatile.Register32
	alarm    [numTimers]volatile.Register32
	armed    volatile.Register32
	timeRawH volatile.Register32
	timeRawL volatile.Register32
	dbgPause volatile.Register32
	pause    volatile.Register32
	intR     volatile.Register32
	intE     volatile.Register32
	intF     volatile.Register32
	intS     volatile.Register32
}

var timer = (*timerType)(unsafe.Pointer(rp.TIMER))

// TimeElapsed returns time elapsed since power up, in microseconds.
func (tmr *timerType) timeElapsed() (us uint64) {
	// Need to make sure that the upper 32 bits of the timer
	// don't change, so read that first
	hi := tmr.timeRawH.Get()
	var lo, nextHi uint32
	for {
		// Read the lower 32 bits
		lo = tmr.timeRawL.Get()
		// Now read the upper 32 bits again and
		// check that it hasn't incremented. If it has, loop around
		// and read the lower 32 bits again to get an accurate value
		nextHi = tmr.timeRawH.Get()
		if hi == nextHi {
			break
		}
		hi = nextHi
	}
	return uint64(hi)<<32 | uint64(lo)
}

// lightSleep will put the processor into a sleep state a short period
// (up to approx 72mins per RP2040 datasheet, 4.6.3. Alarms).
//
// This function is a 'light' sleep and will return early if another
// interrupt or event triggers.  This is intentional since the
// primary use-case is for use by the TinyGo scheduler which will
// re-sleep if needed.
func (tmr *timerType) lightSleep(us uint64) {
	// minSleep is a way to avoid race conditions for short
	// sleeps by ensuring there is enough time to setup the
	// alarm before sleeping.  For very short sleeps, this
	// effectively becomes a 'busy loop'.
	if us < minSleep {
		return
	}

	// Interrupt handler is essentially a no-op, we're just relying
	// on the side-effect of waking the CPU from "wfe"
	intr := interrupt.New(sleepAlarmIRQ, func(interrupt.Interrupt) {
		// Clear the IRQ
		timer.intR.Set(1 << sleepAlarm)
	})

	// Reset interrupt flag
	tmr.intR.Set(1 << sleepAlarm)

	// Enable interrupt
	tmr.intE.SetBits(1 << sleepAlarm)
	intr.Enable()

	// Only the low 32 bits of time can be used for alarms
	target := uint64(tmr.timeRawL.Get()) + us
	tmr.alarm[sleepAlarm].Set(uint32(target))

	// Wait for sleep (or any other) interrupt
	arm.Asm("wfe")

	// Disarm timer
	tmr.armed.Set(1 << sleepAlarm)

	// Disable interrupt
	intr.Disable()
}
