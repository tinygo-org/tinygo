// +build nrf

package runtime

import (
	"device/arm"
	"device/nrf"
)

func _Cfunc_rtc_sleep(ticks uint32)

const Microsecond = 1

//go:export _start
func _start() {
	main()
}

func init() {
	initUART()
	initLFCLK()
	initRTC()
}

func initUART() {
	nrf.UART0.ENABLE = nrf.UART0_ENABLE_ENABLE_Enabled
	nrf.UART0.BAUDRATE = nrf.UART0_BAUDRATE_BAUDRATE_Baud115200
	nrf.UART0.TASKS_STARTTX = 1
	nrf.UART0.PSELTXD = 6 // pin 6 for NRF52840-DK
}

func initLFCLK() {
	nrf.CLOCK.LFCLKSRC = nrf.CLOCK_LFCLKSTAT_SRC_Xtal
	nrf.CLOCK.TASKS_LFCLKSTART = 1
	for nrf.CLOCK.EVENTS_LFCLKSTARTED == 0 {
	}
	nrf.CLOCK.EVENTS_LFCLKSTARTED = 0
}

func initRTC() {
	nrf.RTC0.TASKS_START = 1
	// TODO: set priority
	arm.EnableIRQ(nrf.IRQ_RTC0)
}

func putchar(c byte) {
	nrf.UART0.TXD = nrf.RegValue(c)
	for nrf.UART0.EVENTS_TXDRDY == 0 {
	}
	nrf.UART0.EVENTS_TXDRDY = 0
}

func sleep(d Duration) {
	ticks64 := d / 32
	for ticks64 != 0 {
		monotime()                          // update timestamp
		ticks := uint32(ticks64) & 0x7fffff // 23 bits (to be on the safe side)
		_Cfunc_rtc_sleep(ticks)             // TODO: not accurate (must be d / 30.5175...)
		ticks64 -= Duration(ticks)
	}
}

var (
	timestamp      uint64 // microseconds since boottime
	rtcLastCounter uint32 // 24 bits ticks
)

// Monotonically increasing numer of microseconds since start.
//
// Note: very long pauses between measurements (more than 8 minutes) may
// overflow the counter, leading to incorrect results. This might be fixed by
// handling the overflow event.
func monotime() uint64 {
	rtcCounter := uint32(nrf.RTC0.COUNTER)
	offset := (rtcCounter - rtcLastCounter) % 0xffffff // change since last measurement
	rtcLastCounter = rtcCounter
	timestamp += uint64(offset * 32) // TODO: not precise
	return timestamp
}

func abort() {
	for {
		arm.Asm("wfi")
	}
}

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}
