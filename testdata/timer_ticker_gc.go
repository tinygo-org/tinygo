package main

// Tickers and time.After in many goroutines while the main goroutine runs the GC.
// This used to hang or crash with -scheduler=cores on riscv-qemu, see #6047.

import (
	"runtime"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ticker := time.NewTicker(2 * time.Millisecond)
			defer ticker.Stop()
			timeout := time.After(10 * time.Millisecond)
			for done := false; !done; {
				select {
				case <-ticker.C:
				case <-timeout:
					done = true
				}
			}
		}(i)
	}
	for i := 0; i < 3; i++ {
		runtime.Gosched()
		runtime.GC()
	}
	wg.Wait()
	println("timer ticker gc ok")
}
