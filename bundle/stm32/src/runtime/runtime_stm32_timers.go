//go:build stm32

package runtime

// This file implements a common implementation of implementing 'ticks' and
// 'sleep' for STM32 devices. The implementation uses a single timer.  The
// timer's PWM frequency (controlled by PSC and ARR) are configured for
// periodic interrupts at 100Hz (TICK_INTR_PERIOD_NS).  The PWM counter
// register is used for fine-grained resolution (down to ~150ns) with an
// Output Comparator used for fine-grained sleeps.

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

type timerInfo struct {
	EnableRegister *volatile.Register32
	EnableFlag     uint32
	Device         *stm32.TIM_Type
}

const (
	// All STM32 do a constant 16ns per tick.  This keeps time
	// conversion between ticks and ns fast (shift operation)
	// at the expense of more complex logic when getting current
	// time (which is already slow due to interfacing with hardware)
	NS_PER_TICK = 16

	// For very short sleeps a busy loop is used to avoid race
	// conditions where a trigger would take longer to setup than
	// the sleep duration.
	MAX_BUSY_LOOP_NS = 10e3 // 10us

	// The period of tick interrupts in nanoseconds
	TICK_INTR_PERIOD_NS = 10e6 // 10ms = 100Hz

	// The number of ticks that happen per overflow interrupt
	TICK_PER_INTR = TICK_INTR_PERIOD_NS / NS_PER_TICK
)

var (
	// Tick count since boot
	tickCount volatile.Register64

	// The timer used for counting ticks
	tickTimer *machine.TIM

	// The max counter value (fractional part)
	countMax uint32
)

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * NS_PER_TICK
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / NS_PER_TICK)
}

// number of ticks since start.
//
//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	// For some ways of capturing the time atomically, see this thread:
	// https://www.eevblog.com/forum/microcontrollers/correct-timing-by-timer-overflow-count/msg749617/#msg749617
	// Here, instead of re-reading the counter register if an overflow has been
	// detected, we simply try again because that results in smaller code.
	for {
		mask := interrupt.Disable()
		counter := tickTimer.Count()
		overflows := uint64(tickCount.Get())
		hasOverflow := tickTimer.Device.SR.HasBits(stm32.TIM_SR_UIF)
		interrupt.Restore(mask)

		if hasOverflow {
			continue
		}

		return timeUnit(overflows*TICK_PER_INTR + countToTicks(counter))
	}
}

//
// --  Ticks  ---
//

// Enable the timer used to count ticks.
//
// For precise sleeps use a timer with at least one OutputCompare
// channel otherwise minimum reliable sleep resolution is bounded
// by TICK_INTR_PERIOD_NS.
//
// Typically avoid TIM6 / TIM7 as they often do not include an
// output comparator.
func initTickTimer(tim *machine.TIM) {
	tickTimer = tim
	tickTimer.Configure(machine.PWMConfig{Period: TICK_INTR_PERIOD_NS})

	countMax = tickTimer.Top()
	tickTimer.SetWraparoundInterrupt(handleTick)
}

func handleTick() {
	// increment tick count
	tickCount.Set(tickCount.Get() + 1)
}

//
//  ---  Sleep  ---
//

func sleepTicks(d timeUnit) {
	// If there is a scheduler, we sleep until any kind of CPU event up to
	// a maximum of the requested sleep duration.
	//
	// The scheduler will call again if there is nothing to do and a further
	// sleep is required.
	if hasScheduler {
		timerSleep(uint64(d))
		return
	}

	// There's no scheduler, so we sleep until at least the requested number
	// of ticks has passed.  For short sleeps, this forms a busy loop since
	// timerSleep will return immediately.
	end := ticks() + d
	for ticks() < end {
		timerSleep(uint64(d))
	}
}

// timerSleep sleeps for 'at most' ticks, but possibly less.
func timerSleep(ticks uint64) {
	// If the sleep is super-small (<10us), busy loop by returning
	// to the scheduler (if any).  This avoids a busy loop here
	// that might delay tasks from being scheduled triggered by
	// an interrupt (e.g. channels).
	if ticksToNanoseconds(timeUnit(ticks)) < MAX_BUSY_LOOP_NS {
		return
	}

	// If the sleep is long, the tick interrupt will occur before
	// the sleep expires, so just use that.  This routine will be
	// called again if the sleep is incomplete.
	if ticks >= TICK_PER_INTR {
		waitForEvents()
		return
	}

	// Sleeping for less than one tick interrupt, now see if the
	// next tick interrupt will occur before the sleep expires.  If
	// so, use that interrupt.  (this routine will be called
	// again if sleep is incomplete)
	cnt := tickTimer.Count()
	ticksUntilOverflow := countToTicks(countMax - cnt)
	if ticksUntilOverflow <= ticks {
		waitForEvents()
		return
	}

	// The sleep is now known to be:
	//  - More than a few CPU cycles
	//  - Less than one interrupt period
	//  - Expiring before the next interrupt
	//
	// Setup a PWM channel to trigger an interrupt.
	// NOTE: ticks is known to be < MAX_UINT32 at this point.
	tickTimer.SetMatchInterrupt(0, handleSleep)
	tickTimer.Set(0, cnt+ticksToCount(ticks))

	// Wait till either the timer or some other event wakes
	// up the CPU
	waitForEvents()

	// In case it was not the sleep interrupt that woke the
	// CPU, disable the sleep interrupt now.
	tickTimer.Unset(0)
}

func handleSleep(ch uint8) {
	// Disable the sleep interrupt
	tickTimer.Unset(0)
}

func countToTicks(count uint32) uint64 {
	return (uint64(count) * TICK_PER_INTR) / uint64(countMax)
}

func ticksToCount(ticks uint64) uint32 {
	return uint32((ticks * uint64(countMax)) / TICK_PER_INTR)
}
