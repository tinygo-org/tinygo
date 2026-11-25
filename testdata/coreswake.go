package main

// Test that the cores scheduler wakes cores. QEMU starts the secondary harts
// late, so a test that starts work at once can pass with a broken wakeup.

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"
)

//go:linkname currentCPU runtime.currentCPU
func currentCPU() uint32

func spin(d time.Duration) {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
	}
}

func main() {
	// The sleep lets the other cores go idle first. Then only a wakeup can
	// start the goroutines on them.
	time.Sleep(100 * time.Millisecond)
	var seen [8]atomic.Uint32
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.Now().Add(500 * time.Millisecond)
			for time.Now().Before(deadline) {
				seen[currentCPU()].Store(1)
			}
		}()
	}
	wg.Wait()
	cores := 0
	for i := range seen {
		cores += int(seen[i].Load())
	}
	println("idle cores woken:", cores == runtime.NumCPU())

	// A long sleep owns the timer. A short sleep must interrupt it, not wait
	// for it to end.
	go func() {
		time.Sleep(3 * time.Second)
	}()
	spin(50 * time.Millisecond)
	start := time.Now()
	time.Sleep(20 * time.Millisecond)
	println("short sleep not delayed:", time.Since(start) < 1500*time.Millisecond)

	// Busy every core but one, so the last core sleeps on the timer. main polls a
	// flag, because a blocked main would free its own core for the new goroutine.
	go func() {
		time.Sleep(5 * time.Second)
	}()
	for i := 0; i < runtime.NumCPU()-2; i++ {
		go spin(2 * time.Second)
	}
	spin(50 * time.Millisecond)
	var started atomic.Bool
	start = time.Now()
	go started.Store(true)
	for !started.Load() && time.Since(start) < 2500*time.Millisecond {
	}
	println("sleeping core woken:", started.Load() && time.Since(start) < 500*time.Millisecond)
}
