//go:build rp2040

package main

// This example demonstrates scheduling a delayed interrupt by real time clock.
//
// An interrupt may execute user callback function or used for its side effects
// like waking up from sleep or dormant states.
//
// The interrupt can be configured to repeat.
//
// There is no separate method to disable interrupt, use 0 delay for that.
//
// Unfortunately, it is not possible to use time.Duration to work with RTC directly,
// that would introduce a circular dependency between "machine" and "time" packages.

import (
	"fmt"
	"machine"
	"time"
)

func main() {

	// Schedule and enable recurring interrupt.
	// The callback function is executed in the context of an interrupt handler,
	// so regular restrictions for this sort of code apply: no blocking, no memory allocation, etc.
	// Please check the online documentation for the details about interrupts:
	// https://tinygo.org/docs/concepts/compiler-internals/interrupts/
	delay := time.Minute + 12*time.Second
	machine.RTC.SetInterrupt(uint32(delay.Seconds()), true, func() { println("Peekaboo!") })

	for {
		fmt.Printf("%v\r\n", time.Now().Format(time.RFC3339))
		time.Sleep(1 * time.Second)
	}
}
