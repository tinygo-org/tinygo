//go:build rp2040
// +build rp2040

package main

// This example demonstrates watchdog handling of a runtime panic.
// System hangs on panic, watchdog is not kicked, counter elapses and system reboots.

import (
	"machine"
	"time"
)

var wdc = machine.WatchdogConfig{
	Ticks:        1_100_000, // 1.1 sec
	PauseOnDebug: false,
}

func main() {

	go func() {
		machine.Watchdog.Configure(wdc)
		machine.Watchdog.Enable()
		for {
			machine.Watchdog.Kick()
			time.Sleep(time.Second)
		}
	}()

	i := 10
	for {
		println(i)
		if i == 0 {
			j := 10 / i // simulate runtime panic
			println(j)
		}
		time.Sleep(time.Second)
		i--
	}

}
