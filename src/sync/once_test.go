package sync_test

import (
	"sync"
	"testing"
)

// TestOnceUncontended tests Once on a single goroutine.
func TestOnceUncontended(t *testing.T) {
	var once sync.Once
	{
		var ran bool
		once.Do(func() {
			ran = true
		})
		if !ran {
			t.Error("first call to Do did not run")
		}
	}
	{
		var ran bool
		once.Do(func() {
			ran = true
		})
		if ran {
			t.Error("second call to Do ran")
		}
	}
}

// TestOnceConcurrent tests multiple concurrent invocations of sync.Once.
func TestOnceConcurrent(t *testing.T) {
	var once sync.Once
	var mu sync.Mutex
	mu.Lock()
	var ran bool
	var ranTwice bool
	once.Do(func() {
		ran = true

		// Start a goroutine and (approximately) wait for it to enter the call to Do.
		var startWait sync.Mutex
		startWait.Lock()
		go func() {
			startWait.Unlock()
			once.Do(func() {
				ranTwice = true
			})
			mu.Unlock()
		}()
		startWait.Lock()
	})
	if !ran {
		t.Error("first call to Do did not run")
	}

	// Wait for the goroutine to finish.
	mu.Lock()
	if ranTwice {
		t.Error("second concurrent call to Once also ran")
	}
}
