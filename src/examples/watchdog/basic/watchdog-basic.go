//go:build rp2040
// +build rp2040

package main

// This example demonstrates use of watchdog on RP2040.
// System can be rebooted either by force or timer.
// A reason of the reboot can be extracted too.

import (
	"machine"
	"time"
)

var wdc = machine.WatchdogConfig{
	Ticks:        1_100_000, // 1.1 sec
	PauseOnDebug: false,
}

func main() {

	println()

	wdReason, _ := machine.Watchdog.Reason()
	if wdReason == machine.WatchdogReasonNone {
		println("Hardware reset")
	}
	if wdReason == machine.WatchdogReasonTimer {
		println("Timer reset")
	}
	if wdReason == machine.WatchdogReasonForce {
		println("Force reset")
	}

	machine.Watchdog.Configure(wdc)
	machine.Watchdog.Enable()

	i := 10
	for {
		println(i)
		if i > 0 {
			machine.Watchdog.Kick()
		}
		if i == 5 && wdReason == machine.WatchdogReasonTimer {
			machine.Watchdog.Force()
		}
		time.Sleep(time.Second)
		i--
	}

}
