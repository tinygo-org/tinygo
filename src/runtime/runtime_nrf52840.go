//go:build nrf && nrf52840
// +build nrf,nrf52840

package runtime

import (
	"machine"
	"machine/usb/cdc"
	"runtime/interrupt"
	"runtime/volatile"

	"tinygo.org/x/device/arm"
	"tinygo.org/x/device/nrf"
)

type timeUnit int64

//go:linkname systemInit SystemInit
func systemInit()

//export Reset_Handler
func main() {
	if nrf.FPUPresent {
		arm.SCB.CPACR.Set(0) // disable FPU if it is enabled
	}
	systemInit()
	preinit()
	run()
	exit(0)
}

func init() {
	cdc.EnableUSBCDC()
	machine.USBDev.Configure(machine.UARTConfig{})
	machine.InitSerial()
	initLFCLK()
	initRTC()
}

func initLFCLK() {
	if machine.HasLowFrequencyCrystal {
		nrf.CLOCK.LFCLKSRC.Set(nrf.CLOCK_LFCLKSTAT_SRC_Xtal)
	}
	nrf.CLOCK.TASKS_LFCLKSTART.Set(1)
	for nrf.CLOCK.EVENTS_LFCLKSTARTED.Get() == 0 {
	}
	nrf.CLOCK.EVENTS_LFCLKSTARTED.Set(0)
}

func initRTC() {
	nrf.RTC1.TASKS_START.Set(1)
	intr := interrupt.New(nrf.IRQ_RTC1, func(intr interrupt.Interrupt) {
		if nrf.RTC1.EVENTS_COMPARE[0].Get() != 0 {
			nrf.RTC1.EVENTS_COMPARE[0].Set(0)
			nrf.RTC1.INTENCLR.Set(nrf.RTC_INTENSET_COMPARE0)
			nrf.RTC1.EVENTS_COMPARE[0].Set(0)
			rtc_wakeup.Set(1)
		}
		if nrf.RTC1.EVENTS_OVRFLW.Get() != 0 {
			nrf.RTC1.EVENTS_OVRFLW.Set(0)
			rtcOverflows.Set(rtcOverflows.Get() + 1)
		}
	})
	nrf.RTC1.INTENSET.Set(nrf.RTC_INTENSET_OVRFLW)
	intr.SetPriority(0xc0) // low priority
	intr.Enable()
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
		counter := uint32(nrf.RTC1.COUNTER.Get())
		overflows := rtcOverflows.Get()
		hasOverflow := nrf.RTC1.EVENTS_OVRFLW.Get() != 0
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
	nrf.RTC1.INTENSET.Set(nrf.RTC_INTENSET_COMPARE0)
	rtc_wakeup.Set(0)
	if ticks == 1 {
		// Race condition (even in hardware) at ticks == 1.
		// TODO: fix this in a better way by detecting it, like the manual
		// describes.
		ticks = 2
	}
	nrf.RTC1.CC[0].Set((nrf.RTC1.COUNTER.Get() + ticks) & 0x00ffffff)
	for rtc_wakeup.Get() == 0 {
		waitForEvents()
	}
}
