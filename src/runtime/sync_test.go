//go:build !scheduler.none

package runtime_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSemacquireContention(t *testing.T) {
	var sema uint32 = 1
	var inCrit atomic.Int32
	const goroutines = 4
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				runtime.Semacquire(&sema)
				if inCrit.Add(1) != 1 {
					t.Error("concurrent entry into critical section")
				}
				runtime.Gosched()
				inCrit.Add(-1)
				runtime.Semrelease(&sema)
			}
		}()
	}
	wg.Wait()
}

func TestSemacquireBucketCollision(t *testing.T) {
	// Entries 0 and 16 hash to the same semtable bucket.
	var semas [17]uint32
	var done0, done16 atomic.Bool
	done := make(chan int, 2)
	base := runtime.SemNwait(&semas[0])

	go func() {
		runtime.Semacquire(&semas[0])
		done0.Store(true)
		done <- 0
	}()
	for runtime.SemNwait(&semas[0]) < base+1 {
		runtime.Gosched()
	}

	go func() {
		runtime.Semacquire(&semas[16])
		done16.Store(true)
		done <- 16
	}()
	for runtime.SemNwait(&semas[0]) < base+2 {
		runtime.Gosched()
	}

	runtime.Semrelease(&semas[16])
	if got := <-done; got != 16 {
		t.Fatalf("first wake = %d, want 16", got)
	}
	if done0.Load() {
		t.Fatal("semas[0] waiter woke before release")
	}

	runtime.Semrelease(&semas[0])
	if got := <-done; got != 0 {
		t.Fatalf("second wake = %d, want 0", got)
	}
}
