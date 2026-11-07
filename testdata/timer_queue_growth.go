// Stress test for growing the timer queue while other goroutines sleep and
// allocate. With -scheduler=cores, growing the queue must not allocate while
// schedulerLock is held. A GC started there waits for a core that is spinning
// on the lock with interrupts off, and the program hangs.
package main

import (
	"runtime"
	"time"
)

func main() {
	// Contend on schedulerLock via short sleeps.
	go func() {
		for {
			time.Sleep(100 * time.Microsecond)
		}
	}()
	// Allocation churn keeps the heap close to the GC threshold, so the next
	// allocation (such as a queue grow) is likely to start a GC.
	for i := 0; i < 2; i++ {
		go func() {
			var sink []byte
			for {
				sink = make([]byte, 4096)
				sink[0] = 1
			}
		}()
	}

	// Grow the timer queue in rounds. Each round starts from nothing so the
	// slice has to grow (and allocate under the lock) again.
	for round := 0; round < 10; round++ {
		timers := make([]*time.Timer, 0, 1000)
		for i := 0; i < 1000; i++ {
			timers = append(timers, time.NewTimer(time.Hour))
		}
		for _, t := range timers {
			t.Stop()
		}
		runtime.GC()
	}
	println("timer queue growth ok")
}
