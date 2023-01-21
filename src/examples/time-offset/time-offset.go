package main

// This example demonstrates how to set the system time.
//
// Usually, boards don't keep time on power cycles and restarts, it resets to Unix epoch.
// The system time can be set by calling runtime.AdjustTimeOffset().
//
// Possible time sources: an external RTC clock or network (WiFi access point or NTP server)

import (
	"fmt"
	"runtime"
	"time"
)

const myTime = "2006-01-02T15:04:05Z" // this is an example time you source somewhere

func main() {

	// measure how many nanoseconds the internal clock is behind
	timeOfMeasurement := time.Now()
	actualTime, _ := time.Parse(time.RFC3339, myTime)
	offset := actualTime.Sub(timeOfMeasurement)

	// adjust internal clock by adding the offset to the internal clock
	runtime.AdjustTimeOffset(int64(offset))

	for {
		fmt.Printf("%v\r\n", time.Now().Format(time.RFC3339))
		time.Sleep(5 * time.Second)
	}
}
