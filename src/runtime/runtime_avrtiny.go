//go:build avrtiny

// Runtime for the newer AVRs introduced since around 2016 that work quite
// different from older AVRs like the atmega328p or even the attiny85.
// Because of these large differences, a new runtime and machine implementation
// is needed.
// Some key differences:
//   * Peripherals are now logically separated, instead of all mixed together as
//     one big bag of registers. No PORTA/DDRA etc registers anymore, instead a
//     real PORT peripheral type with multiple instances.
//   * There is a real RTC now! No need for using one of the timers as a time
//     source, which then conflicts with using it as a PWM.
//   * Flash and RAM are now in the same address space! This avoids the need for
//     PROGMEM which couldn't (easily) be supported in Go anyway. Constant
//     globals just get stored in flash, like on Cortex-M chips.

package runtime

import (
	"device/avr"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

//export main
func main() {
	// Initialize RTC.
	for avr.RTC.STATUS.Get() != 0 {
	}
	avr.RTC.CTRLA.Set(avr.RTC_CTRLA_RTCEN | avr.RTC_CTRLA_RUNSTDBY)
	avr.RTC.INTCTRL.Set(avr.RTC_INTCTRL_OVF) // enable overflow interrupt
	interrupt.New(avr.IRQ_RTC_CNT, rtcInterrupt)

	// Configure sleep:
	// - enable sleep mode
	// - set sleep mode to STANDBY (mode 0x1)
	avr.SLPCTRL.CTRLA.Set(avr.SLPCTRL_CTRLA_SEN | 0x1<<1)

	// Enable interrupts after initialization.
	avr.Asm("sei")

	run()
	exit(0)
}

func initUART() {
	// no UART configured
}

func putchar(b byte) {
	// no-op
}

// ticksToNanoseconds converts RTC ticks (at 32768Hz) to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ticks * 1e9 / 32768
	return int64(ticks) * 1953125 / 64
}

// nanosecondsToTicks converts nanoseconds to RTC ticks (running at 32768Hz).
func nanosecondsToTicks(ns int64) timeUnit {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ns * 32768 / 1e9
	return timeUnit(ns * 64 / 1953125)
}

// Sleep for the given number of timer ticks.
func sleepTicks(d timeUnit) {
	ticksStart := ticks()
	sleepUntil := ticksStart + d

	// Sleep until we're in the right 2-second interval.
	for {
		avr.Asm("cli")
		overflows := rtcOverflows.Get()
		if overflows >= uint32(sleepUntil>>16) {
			// We're in the right 2-second interval.
			// At this point we know that the difference between ticks() and
			// sleepUntil is ≤0xffff.
			avr.Asm("sei")
			break
		}
		// Sleep some more, because we're not there yet.
		avr.Asm("sei\nsleep")
	}

	// Now we know the sleep duration is small enough to fit in rtc.CNT.

	// Update rtc.CMP (atomically).
	cnt := uint16(sleepUntil)
	low := uint8(cnt)
	high := uint8(cnt >> 8)
	avr.RTC.CMPH.Set(high)
	avr.RTC.CMPL.Set(low)

	// Disable interrupts, so we can change interrupt settings without racing.
	avr.Asm("cli")

	// Enable the CMP interrupt.
	avr.RTC.INTCTRL.Set(avr.RTC_INTCTRL_OVF | avr.RTC_INTCTRL_CMP)

	// Check whether we already reached CNT, in which case the interrupt may
	// have triggered already (but maybe not, it's a race condition).
	low2 := avr.RTC.CNTL.Get()
	high2 := avr.RTC.CNTH.Get()
	cnt2 := uint16(high2)<<8 | uint16(low2)
	if cnt2 < cnt {
		// We have not, so wait until the interrupt happens.
		for {
			// Sleep until the next interrupt happens.
			avr.Asm("sei\nsleep\ncli")
			if cmpMatch.Get() != 0 {
				// The CMP interrupt occured, so we have slept long enough.
				cmpMatch.Set(0)
				break
			}
		}
	}

	// Disable the CMP interrupt, and restore things like they were before.
	avr.RTC.INTCTRL.Set(avr.RTC_INTCTRL_OVF)
	avr.Asm("sei")
}

// Number of RTC overflows, updated in the RTC interrupt handler.
// The RTC is running at 32768Hz so an overflow happens every 2 seconds. A
// 32-bit integer is large enough to run for about 279 years.
var rtcOverflows volatile.Register32

// Set to one in the RTC CMP interrupt, to signal the expected number of ticks
// have passed.
var cmpMatch volatile.Register8

// Return the number of RTC ticks that happened since reset.
func ticks() timeUnit {
	var ovf uint32
	var count uint16
	for {
		// Get the tick count and overflow value, in a 4-step process to avoid a
		// race with the overflow interrupt.
		mask := interrupt.Disable()

		// 1. Get the overflow value.
		ovf = rtcOverflows.Get()

		// 2. Read the RTC counter.
		// This way of reading is atomic (due to the TEMP register).
		low := avr.RTC.CNTL.Get()
		high := avr.RTC.CNTH.Get()

		// 3. Get the interrupt flags.
		intflags := avr.RTC.INTFLAGS.Get()

		interrupt.Restore(mask)

		// 4. Check whether an overflow happened somewhere in the last three
		// steps. If so, just repeat the loop.
		if intflags&avr.RTC_INTFLAGS_OVF == 0 {
			count = uint16(high)<<8 | uint16(low)
			break
		}
	}

	// Create the 64-bit tick count, combining the two.
	return timeUnit(ovf)<<16 | timeUnit(count)
}

// Interrupt handler for the RTC.
// It happens every two seconds, and while sleeping using the CMP interrupt.
func rtcInterrupt(interrupt.Interrupt) {
	flags := avr.RTC.INTFLAGS.Get()
	if flags&avr.RTC_INTFLAGS_OVF != 0 {
		rtcOverflows.Set(rtcOverflows.Get() + 1)
	}
	if flags&avr.RTC_INTFLAGS_CMP != 0 {
		cmpMatch.Set(1)
	}
	avr.RTC.INTFLAGS.Set(flags) // clear interrupts
}

func exit(code int) {
	abort()
}

//export __vector_default
func abort()
