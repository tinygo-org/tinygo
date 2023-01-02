//go:build nrf

package runtime

import (
	"device/nrf"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

var rtcHandle *nrf.RTC_Type

//go:linkname systemInit SystemInit
func systemInit()

func initRTC(rtc *nrf.RTC_Type) {
	rtcHandle = rtc
	rtcHandle.TASKS_START.Set(1)
	intr := interrupt.New(nrf.IRQ_RTC1, func(intr interrupt.Interrupt) {
		if rtcHandle.EVENTS_COMPARE[0].Get() != 0 {
			rtcHandle.EVENTS_COMPARE[0].Set(0)
			rtcHandle.INTENCLR.Set(nrf.RTC_INTENSET_COMPARE0)
			rtcHandle.EVENTS_COMPARE[0].Set(0)
			rtc_wakeup.Set(1)
		}
		if rtcHandle.EVENTS_OVRFLW.Get() != 0 {
			rtcHandle.EVENTS_OVRFLW.Set(0)
			rtcOverflows.Set(rtcOverflows.Get() + 1)
		}
	})
	rtcHandle.INTENSET.Set(nrf.RTC_INTENSET_OVRFLW)
	intr.SetPriority(0xc0) // low priority
	intr.Enable()
}

func putchar(c byte) {
	machine.Serial.WriteByte(c)
}

func getchar() byte {
	for buffered() == 0 {
		Gosched()
	}
	v, _ := machine.Serial.ReadByte()
	return v
}

func buffered() int {
	return machine.Serial.Buffered()
}

func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks := uint32(d) & 0x7fffff // 23 bits (to be on the safe side)
		rtc_sleep(ticks)
		d -= timeUnit(ticks)
	}
}

var rtcOverflows volatile.Register32 // number of times the RTC wrapped around

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

// Monotonically increasing numer of ticks since start.
func ticks() timeUnit {
	// For some ways of capturing the time atomically, see this thread:
	// https://www.eevblog.com/forum/microcontrollers/correct-timing-by-timer-overflow-count/msg749617/#msg749617
	// Here, instead of re-reading the counter register if an overflow has been
	// detected, we simply try again because that results in (slightly) smaller
	// code and is perhaps easier to prove correct.
	for {
		mask := interrupt.Disable()
		counter := uint32(rtcHandle.COUNTER.Get())
		overflows := rtcOverflows.Get()
		hasOverflow := rtcHandle.EVENTS_OVRFLW.Get() != 0
		interrupt.Restore(mask)

		if hasOverflow {
			// There was an overflow. Try again.
			continue
		}

		// The counter is 24 bits in size, so the number of overflows form the
		// upper 32 bits (together 56 bits, which covers 71493 years at
		// 32768kHz: I'd argue good enough for most purposes).
		return timeUnit(overflows)<<24 + timeUnit(counter)
	}
}

var rtc_wakeup volatile.Register8

func rtc_sleep(ticks uint32) {
	rtcHandle.INTENSET.Set(nrf.RTC_INTENSET_COMPARE0)
	rtc_wakeup.Set(0)
	if ticks == 1 {
		// Race condition (even in hardware) at ticks == 1.
		// TODO: fix this in a better way by detecting it, like the manual
		// describes.
		ticks = 2
	}
	rtcHandle.CC[0].Set((rtcHandle.COUNTER.Get() + ticks) & 0x00ffffff)
	for rtc_wakeup.Get() == 0 {
		waitForEvents()
	}
}
