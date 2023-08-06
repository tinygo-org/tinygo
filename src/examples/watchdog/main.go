package main

import (
	"fmt"
	"machine"
	"time"
)

func main() {
	//sleep for 2 secs for console
	time.Sleep(2 * time.Second)

	config := machine.WatchdogConfig{
		TimeoutMillis: 1000,
	}

	println("configuring watchdog for max 1 second updates")
	machine.Watchdog.Configure(config)

	// From this point the watchdog is running and Update must be
	// called periodically, per the config
	machine.Watchdog.Start()

	// This loop should complete because watchdog update is called
	// every 100ms.
	start := time.Now()
	println("updating watchdog for 3 seconds")
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		machine.Watchdog.Update()
		fmt.Printf("alive @ %v\n", time.Now().Sub(start))
	}

	// This loop should cause a watchdog reset after 1s since
	// there is no update call.
	start = time.Now()
	println("entering tight loop")
	for {
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("alive @ %v\n", time.Now().Sub(start))
	}
}
