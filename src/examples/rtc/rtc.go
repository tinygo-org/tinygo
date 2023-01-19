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

	// Set RTC from system time and enable recurring alarm
	machine.RTC.SetTime(toRtcTime(time.Now()))
	machine.RTC.SetAlarm(rtcTimeAlarm, func() { println("Pekabo!") }) // the callback function executes on interrupt and shall be as quick as possible

	// Wait a bit to let user connect to serial console and for RTC to initialize
	time.Sleep(3 * time.Second)

	for {
		rtcTime, _ := machine.RTC.GetTime() // shall increase 1 second each time
		nowRtc := toTime(rtcTime)
		nowSys := time.Now()
		diff := nowRtc.Sub(nowSys)
		// Some descrepancy is expected, diff shall be constant
		fmt.Printf("SYS: %v, RTC: %v, DIFF: %v\r\n", nowSys.Format(time.RFC3339), nowRtc.Format(time.RFC3339), diff)
		time.Sleep(time.Second)
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
