// Several goroutines grow and drain the timer queue at the same time, and
// every short timer must fire exactly once. With -scheduler=cores this also
// covers two cores growing the queue at once.
package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	workers = 6
	rounds  = 10
	longN   = 800
	shortN  = 50
)

var fired atomic.Uint64

func main() {
	stop := make(chan struct{})
	go func() {
		var sink []byte
		for {
			select {
			case <-stop:
				return
			default:
			}
			sink = make([]byte, 16384)
			sink[0] = 1
			runtime.Gosched()
		}
	}()

	for round := 0; round < rounds; round++ {
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				long := make([]*time.Timer, 0, longN)
				for i := 0; i < longN; i++ {
					long = append(long, time.NewTimer(time.Hour))
				}
				for i := 0; i < shortN; i++ {
					time.AfterFunc(time.Duration(1+i%5)*time.Millisecond, func() { fired.Add(1) })
				}
				for _, t := range long {
					t.Stop()
				}
			}()
		}
		wg.Wait()
		want := uint64(workers * shortN * (round + 1))
		deadline := time.Now().Add(5 * time.Second)
		for fired.Load() != want {
			if time.Now().After(deadline) {
				println("FAIL round", round, "fired", fired.Load(), "want", want)
				return
			}
			time.Sleep(time.Millisecond)
		}
		runtime.GC()
	}
	close(stop)
	println("concurrent timer growth ok")
}
