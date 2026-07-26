package main

import (
	"machine"
	"time"
)

func main() {
	time.Sleep(2 * time.Second)

	println("configuring window watchdog")
	config := machine.WindowWatchdogConfig{
		TimeoutMicros: 100000, // 100ms
		WindowPercent: 50,     // 50ms to 100ms refresh window
	}

	machine.WindowWatchdog.Configure(config)
	machine.WindowWatchdog.Start()

	println("updating wwdg for 1 second")
	for i := 0; i < 10; i++ {
		time.Sleep(75 * time.Millisecond) // middle of the window
		machine.WindowWatchdog.Update()
		println("alive")
	}

	println("entering tight loop (will reset)")
	for {
		time.Sleep(10 * time.Millisecond)
	}
}
