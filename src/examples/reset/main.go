// Prints why the chip last reset, then resets it with machine.CPUReset.
package main

import (
	"machine"
	"time"
)

func main() {
	time.Sleep(2 * time.Second)

	println("reset reason:", machine.ResetReason().String())
	println("reset reason raw:", machine.ResetReasonRaw())

	println("resetting in 3 seconds")
	time.Sleep(3 * time.Second)
	machine.CPUReset()
}
