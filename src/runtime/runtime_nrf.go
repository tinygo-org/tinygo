// +build nrf

package runtime

import (
	"device/arm"
	"device/nrf"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
)

type timeUnit int64

const tickMicros = 1024 * 32

//go:linkname systemInit SystemInit
func systemInit()

//go:export Reset_Handler
func main() {
	systemInit()
	preinit()
	initAll()
	callMain()
	abort()
}

func init() {
	machine.UART0.Configure(machine.UARTConfig{})
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
		nrf.RTC1.INTENCLR.Set(nrf.RTC_INTENSET_COMPARE0)
		nrf.RTC1.EVENTS_COMPARE[0].Set(0)
		rtc_wakeup.Set(1)
	})
	intr.SetPriority(0xc0) // low priority
	intr.Enable()
}

func putchar(c byte) {
	machine.UART0.WriteByte(c)
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	for d != 0 {
		ticks()                       // update timestamp
		ticks := uint32(d) & 0x7fffff // 23 bits (to be on the safe side)
		rtc_sleep(ticks)              // TODO: not accurate (must be d / 30.5175...)
		d -= timeUnit(ticks)
	}
}

var (
	timestamp      timeUnit // nanoseconds since boottime
	rtcLastCounter uint32   // 24 bits ticks
)

// Monotonically increasing numer of ticks since start.
//
// Note: very long pauses between measurements (more than 8 minutes) may
// overflow the counter, leading to incorrect results. This might be fixed by
// handling the overflow event.
func ticks() timeUnit {
	rtcCounter := uint32(nrf.RTC1.COUNTER.Get())
	offset := (rtcCounter - rtcLastCounter) & 0xffffff // change since last measurement
	rtcLastCounter = rtcCounter
	timestamp += timeUnit(offset) // TODO: not precise
	return timestamp
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
		arm.Asm("wfi")
	}
}
