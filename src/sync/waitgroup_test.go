package sync_test

import (
	"sync"
	"testing"
)

// TestWaitGroupUncontended tests the wait group from a single goroutine.
func TestWaitGroupUncontended(t *testing.T) {
	// Check that a single add-and-done works.
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
	wg.Wait()

	// Check that mixing positive and negative counts works.
	wg.Add(10)
	wg.Add(-8)
	wg.Add(-1)
	wg.Add(0)
	wg.Done()
	wg.Wait()
}

// TestWaitGroup tests the typical usage of WaitGroup.
func TestWaitGroup(t *testing.T) {
	const n = 5
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go wg.Done()
	}

	wg.Wait()
}
