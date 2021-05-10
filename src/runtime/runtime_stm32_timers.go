// +build stm32

package runtime

// This file implements a common implementation of implementing 'ticks' and 'sleep' for STM32 devices. The
// implementation uses two 'basic' timers, so should be compatible with a broad range of STM32 MCUs.
//
// This implementation is of 'sleep' is for running in a normal power mode.  Use of the RTC to enter and
// resume from low-power states is out of scope.
//
// Interface
// ---------
// For each MCU, the following constants should be defined:
//   TICK_RATE             The desired rate of ticks, e.g. 10000 for 10KHz ticks
//   TICK_TIMER_INTFREQ    The desired frequency of timer interrupts ex: 1000 for 1Khz interrupts
//   TICK_TIMER_IRQ        Which timer to use for counting ticks (e.g. stm32.IRQ_TIM7)
//   TICK_TIMER_CLOCKFREQ  The frequency the clock feeding the sleep timer is set to (e.g. 84MHz)
//   SLEEP_TIMER_IRQ       Which timer to use for sleeping (e.g. stm32.IRQ_TIM3)
//   SLEEP_TIMER_CLOCKFREQ The frequency the clock feeding the sleep timer is set to (e.g. 84MHz)
//
// The type alias `arrtype` should be defined to either uint32 or uint16 depending on the
// size of that register in the MCU's TIM_Type structure

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
)

type timerInfo struct {
	EnableRegister *volatile.Register32
	EnableFlag     uint32
	Device         *stm32.TIM_Type
}

const (
	TICKS_PER_NS   = 1000000000 / TICK_RATE
	TICKS_PER_INTS = TICK_RATE / TICK_TIMER_INTFREQ
)

var (
	// Tick count since boot
	tickCount volatile.Register64

	// The timer used for counting ticks
	tickTimer *timerInfo

	// The timer used for sleeping
	sleepTimer *timerInfo

	// timer Auto-reload value
	timerARR uint32
)

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * TICKS_PER_NS
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / TICKS_PER_NS)
}

// number of ticks (microseconds) since start.
//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	// Global tickCount variable is updated at every interrupt, but this is a coarse estimation
	// of elapsed time. (tickCount never includes elapsed time between two interrupts)
	// In this function, we need a precise value of elapsed time.
	// This is the reason we complement tickCount with extra time computed from timer's current value.
	fractional := (TICKS_PER_INTS * tickTimer.Device.CNT.Get()) / timerARR
	preciseTickCount := tickCount.Get() + uint64(fractional)
	return timeUnit(preciseTickCount)
}

//
// --  Ticks  ---
//

// Enable the timer used to count ticks
func initTickTimer(ti *timerInfo) {

	tickTimer = ti
	ti.EnableRegister.SetBits(ti.EnableFlag)

	// Compute pre-scaler
	timerPSC := uint32(TICK_TIMER_CLOCKFREQ / TICK_TIMER_INTFREQ)
	timerARR = uint32(1)

	// Increase Auto Reload as much as possible
	for timerPSC > 1 && timerARR < 0x7FFF {
		timerPSC >>= 1
		timerARR <<= 1
	}
	ti.Device.PSC.Set(timerPSC - 1)
	ti.Device.ARR.Set(arrtype(timerARR - 1))

	// Auto-repeat
	ti.Device.EGR.SetBits(stm32.TIM_EGR_UG)

	// Register the interrupt handler
	intr := interrupt.New(TICK_TIMER_IRQ, handleTick)
	intr.SetPriority(0xc1)
	intr.Enable()

	// Clear update flag
	ti.Device.SR.ClearBits(stm32.TIM_SR_UIF)

	// Enable the hardware interrupt
	ti.Device.DIER.SetBits(stm32.TIM_DIER_UIE)

	// Enable the timer
	ti.Device.CR1.SetBits(stm32.TIM_CR1_CEN)
}

func handleTick(interrupt.Interrupt) {

	if tickTimer.Device.SR.HasBits(stm32.TIM_SR_UIF) {
		// clear the update flag
		tickTimer.Device.SR.ClearBits(stm32.TIM_SR_UIF)

		// increment tick count (coarse value)
		tickCount.Set(tickCount.Get() + TICKS_PER_INTS)
	}
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
		timerSleep(ticksToNanoseconds(d))
		return
	}

	// There's no scheduler, so we sleep until at least the requested number
	// of ticks has passed.
	end := ticks() + d
	for ticks() < end {
		timerSleep(ticksToNanoseconds(d))
	}
}

// Enable the Sleep clock
func initSleepTimer(ti *timerInfo) {
	sleepTimer = ti

	ti.EnableRegister.SetBits(ti.EnableFlag)

	// No auto-repeat
	ti.Device.EGR.SetBits(stm32.TIM_EGR_UG)

	// Enable the hardware interrupt.
	ti.Device.DIER.SetBits(stm32.TIM_DIER_UIE)

	intr := interrupt.New(SLEEP_TIMER_IRQ, handleSleep)
	intr.SetPriority(0xc3)
	intr.Enable()
}

// timerSleep sleeps for 'at most' ns nanoseconds, but possibly less.
func timerSleep(ns int64) {
	// Calculate initial pre-scale value.
	// delay (in ns) and clock freq are both large values, so do the nanosecs
	// conversion (divide by 1G) by pre-dividing each by 1000 to avoid overflow
	// in any meaningful time period.
	psc := ((ns / 1000) * (SLEEP_TIMER_CLOCKFREQ / 1000)) / 1000
	period := int64(1)

	// Get the pre-scale into range, with interrupt firing
	// once per tick.
	for psc > 0x10000 || period == 1 {
		psc >>= 1
		period <<= 1
	}

	// Clamp overflow
	if period > 0x10000 {
		period = 0x10000
	}

	// Set the desired duration and enable
	sleepTimer.Device.PSC.Set(uint32(psc) - 1)
	sleepTimer.Device.ARR.Set(arrtype(period) - 1)
	sleepTimer.Device.CR1.SetBits(stm32.TIM_CR1_CEN)

	// Wait till either the timer or some other event wakes
	// up the CPU
	waitForEvents()

	// In case it was not the sleep timer that woke the
	// CPU, disable the timer now.
	disableSleepTimer()
}

func handleSleep(interrupt.Interrupt) {
	disableSleepTimer()
}

func disableSleepTimer() {
	// Disable and clear the update flag.
	sleepTimer.Device.CR1.ClearBits(stm32.TIM_CR1_CEN)
	sleepTimer.Device.SR.ClearBits(stm32.TIM_SR_UIF)
}
