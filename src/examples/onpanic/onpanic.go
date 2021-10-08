// +build cortexm

package main

import (
	"device/arm"
	"runtime"
	"time"
)

// This example shows how to reset system on panic.
// In your application you may: reset, blink an LED to indicate failure, or do something else.

func main() {

	runtime.OnPanic = func() {
		println("PANIC")
		time.Sleep(time.Second)
		arm.SystemReset() // available on cortexm, your system may have to do this differently
	}

	println("START")
	time.Sleep(time.Second)
	panic("AAA!!!111")

}
