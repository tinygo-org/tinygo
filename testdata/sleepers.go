package main

// Stress test for time.Sleep and timers with many goroutines.

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	numSleepers  = 16
	maxOvershoot = time.Second
)

func main() {
	var failures atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < numSleepers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				d := time.Duration(10+(i*7)%40) * time.Millisecond
				start := time.Now()
				time.Sleep(d)
				elapsed := time.Since(start)
				if elapsed < d {
					println("sleeper", i, "woke early:", elapsed.String(), "<", d.String())
					failures.Add(1)
				} else if elapsed > d+maxOvershoot {
					println("sleeper", i, "overslept:", elapsed.String(), ">", d.String())
					failures.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()

	time.Sleep(0)
	time.Sleep(-time.Second)
	time.Sleep(1)

	start := time.Now()
	ticker := time.NewTicker(time.Millisecond)
	for i := 0; i < 100; i++ {
		<-ticker.C
	}
	ticker.Stop()
	if elapsed := time.Since(start); elapsed < 99*time.Millisecond {
		println("ticker ran too fast:", elapsed.String())
		failures.Add(1)
	}

	start = time.Now()
	<-time.After(20 * time.Millisecond)
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		println("time.After fired early:", elapsed.String())
		failures.Add(1)
	}

	if failures.Load() != 0 {
		panic("sleep failures")
	}
	println("sleepers ok")
}
