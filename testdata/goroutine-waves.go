package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Many goroutines that start and end together. This reuses the threads of the
// finished goroutines and ends the ones that are not needed.

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

var (
	waves    = 40
	perWave  = 100
	exitRuns = 20
)

// base is the goroutine count when main starts, which is not 1 on every target.
var base int

func waitForBase() {
	for i := 0; i < 1000 && runtime.NumGoroutine() > base; i++ {
		time.Sleep(time.Millisecond)
	}
}

func main() {
	base = runtime.NumGoroutine()
	if small {
		waves, perWave, exitRuns = 8, 20, 10
	}

	// Waves larger than the pool of idle threads.
	var total atomic.Int64
	for w := 0; w < waves; w++ {
		var wg sync.WaitGroup
		for i := 0; i < perWave; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				total.Add(int64(i + 1))
				runtime.Gosched()
			}()
		}
		wg.Wait()
	}
	want := int64(waves * perWave * (perWave + 1) / 2)
	println("waves sum ok:", total.Load() == want)
	waitForBase()
	println("goroutines after waves:", runtime.NumGoroutine()-base)

	// Goexit more often than a thread can restart, with deferred calls.
	var exited atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < exitRuns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer exited.Add(1)
			func() {
				defer runtime.Goexit()
				runtime.Gosched()
			}()
			println("not reached")
		}()
		wg.Wait()
	}
	println("goexit runs:", exited.Load() == int64(exitRuns))

	// A recovered panic must not leave state in a reused thread.
	var recovered atomic.Int64
	for i := 0; i < exitRuns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if recover() != nil {
					recovered.Add(1)
				}
			}()
			var m map[int]int
			m[1] = 1
		}()
		wg.Wait()
	}
	println("recovered panics:", recovered.Load() == int64(exitRuns))

	// Goroutines that block before ending, all woken at once.
	start := make(chan struct{})
	var ran atomic.Int64
	for i := 0; i < perWave; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ran.Add(1)
		}()
	}
	runtime.Gosched()
	close(start)
	wg.Wait()
	println("woken:", ran.Load() == int64(perWave))
	waitForBase()
	println("goroutines at end:", runtime.NumGoroutine()-base)
}
