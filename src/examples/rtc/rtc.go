//go:build rp2040

package main

// This example demonstrates real time clock (RTC) support.
//
// An alarm can be set to execute user callback function once or on a schedule.
//
// Unfortunately, it is not possible to use time.Time to work with RTC,
// that would introduce a circular dependency between "machine" and "time" packages.

import (
	"fmt"
	"machine"
	"time"
)

// For RTC to work, it must be set to some reference time first
var rtcTimeReference = machine.RtcTime{
	Year:  2023,
	Month: 01,
	Day:   19,
	Dotw:  4,
	Hour:  1,
	Min:   12,
	Sec:   45,
}

// Alarm shall fire every minute at 5 sec
var rtcTimeAlarm = machine.RtcTime{
	Year:  -1,
	Month: -1,
	Day:   -1,
	Dotw:  -1,
	Hour:  -1,
	Min:   -1,
	Sec:   5,
}

func main() {

	// configure RTC and set alarm
	machine.RTC.SetTime(rtcTimeReference)
	machine.RTC.SetAlarm(rtcTimeAlarm, func() { println("Pekabo!") }) // the callback function executes on interrupt and shall be as quick as possible

	// wait a bit to let user connect to serial console and for RTC to initialize
	time.Sleep(3 * time.Second)

	for {
		rtcTime, err := machine.RTC.GetTime() // reading time from RTC, it shall increase 1 second each read
		if err != nil {
			println(err.Error())
		}
		printTime(rtcTime)
		time.Sleep(time.Second)
	}
}

func printTime(t machine.RtcTime) {
	fmt.Printf("%4d-%02d-%02d %s %02d:%02d:%02d\r\n", t.Year, t.Month, t.Day, time.Weekday(t.Dotw).String()[:3], t.Hour, t.Min, t.Sec)
}
