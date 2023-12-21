package spinlock

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func testLock(threads, n int, l sync.Locker) time.Duration {
	var wg sync.WaitGroup
	wg.Add(threads)

	var count1 int
	var count2 int

	start := time.Now()
	for i := 0; i < threads; i++ {
		go func() {
			for i := 0; i < n; i++ {
				l.Lock()
				count1++
				count2 += 2
				l.Unlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	dur := time.Since(start)
	if count1 != threads*n {
		panic("mismatch")
	}
	if count2 != threads*n*2 {
		panic("mismatch")
	}
	return dur
}

func TestSpinLock(t *testing.T) {
	fmt.Printf("[1] spinlock %4.0fms\n", testLock(1, 1000000, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[1] mutex    %4.0fms\n", testLock(1, 1000000, &sync.Mutex{}).Seconds()*1000)
	fmt.Printf("[4] spinlock %4.0fms\n", testLock(4, 1000000, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[4] mutex    %4.0fms\n", testLock(4, 1000000, &sync.Mutex{}).Seconds()*1000)
	fmt.Printf("[8] spinlock %4.0fms\n", testLock(8, 1000000, &SpinLock{}).Seconds()*1000)
	fmt.Printf("[8] mutex    %4.0fms\n", testLock(8, 1000000, &sync.Mutex{}).Seconds()*1000)
}
