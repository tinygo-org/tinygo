// Goroutines on several cores grow the timer queue while the GC runs.
// Every goroutine must keep seeing its own slices and timers.
package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	growers   = 3
	hammerers = 3
	rounds    = 12
	perRound  = 120
)

var (
	fired atomic.Uint64
	stop  atomic.Bool
)

func main() {
	var bg sync.WaitGroup

	bg.Add(1)
	go func() {
		defer bg.Done()
		var sink []byte
		for i := 0; !stop.Load(); i++ {
			sink = make([]byte, 8192)
			sink[0] = 1
			if i%20 == 0 {
				runtime.GC()
			}
			runtime.Gosched()
		}
	}()

	for i := 0; i < hammerers; i++ {
		bg.Add(1)
		go func() {
			defer bg.Done()
			for !stop.Load() {
				t := time.NewTimer(time.Hour)
				t.Stop()
				runtime.Gosched()
			}
		}()
	}

	var keep [growers][]*time.Timer
	for round := 0; round < rounds; round++ {
		var wg sync.WaitGroup
		for g := 0; g < growers; g++ {
			g := g
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < perRound; i++ {
					keep[g] = append(keep[g], time.NewTimer(time.Hour))
				}
				for i := 0; i < 5; i++ {
					time.AfterFunc(time.Duration(1+i)*time.Millisecond, func() { fired.Add(1) })
				}
			}()
		}
		wg.Wait()
		want := uint64(growers * 5 * (round + 1))
		deadline := time.Now().Add(10 * time.Second)
		for fired.Load() != want {
			if time.Now().After(deadline) {
				println("FAIL round", round, "fired", fired.Load(), "want", want)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}

	stop.Store(true)
	bg.Wait()
	for _, ts := range keep {
		for _, t := range ts {
			t.Stop()
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.NumGC == 0 {
		println("FAIL no GC ran")
		return
	}
	println("timer growth under gc ok")
}
