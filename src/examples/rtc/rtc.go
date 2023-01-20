//go:build rp2040

package main

// This example demonstrates real time clock (RTC) support.
//
// An alarm can be set to execute user callback function once or on a schedule.
//
// Unfortunately, it is not possible to use time.Time to work with RTC directly,
// that would introduce a circular dependency between "machine" and "time" packages.

import (
	"fmt"
	"machine"
	"runtime"
	"time"
)

// Alarm shall fire every minute at 20 sec
var rtcTimeAlarm = machine.RtcTime{
	Year:  -1,
	Month: -1,
	Day:   -1,
	Dotw:  -1,
	Hour:  -1,
	Min:   -1,
	Sec:   20,
}

func main() {

	// Set system clock to some date and time.
	t, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	runtime.AdjustTimeOffset(-1 * int64(time.Since(t)))

	// Set RTC from system time.
	// Some descrepancy is expected. RTC may show 1 second ahead system time.
	machine.RTC.SetTime(toRtcTime(time.Now()))

	// Schedule and enable recurring alarm.
	// The callback function is executed in the context of an interrupt handler,
	// so regular restructions for this sort of code apply: no blocking, no memory allocation, etc.
	machine.RTC.SetAlarm(rtcTimeAlarm, func() { println("Pekabo!") })

	for {
		rtcTime, _ := machine.RTC.GetTime()
		nowRtc := toTime(rtcTime)
		nowSys := time.Now()

		// One iterration takes slightly more than a second.
		// Both clocks run from the same source internally (XOSC) and are in sync.
		// Since we don't show milliseconds, the accumulated error remains invisible for roughly 20 minutes.
		// Then both SYS and RTC values will "jump" two seconds a time.
		fmt.Printf("SYS: %v, RTC: %v\r\n", nowSys.Format(time.RFC3339), nowRtc.Format(time.RFC3339))

		time.Sleep(1 * time.Second)
	}
}

func toRtcTime(t time.Time) machine.RtcTime {
	return machine.RtcTime{
		Year:  int16(t.Year()),
		Month: int8(t.Month()),
		Day:   int8(t.Day()),
		Dotw:  int8(t.Weekday()),
		Hour:  int8(t.Hour()),
		Min:   int8(t.Minute()),
		Sec:   int8(t.Second()),
	}
}

func toTime(t machine.RtcTime) time.Time {
	return time.Date(
		int(t.Year),
		time.Month(t.Month),
		int(t.Day),
		int(t.Hour),
		int(t.Min),
		int(t.Sec),
		0, time.Local)
}
