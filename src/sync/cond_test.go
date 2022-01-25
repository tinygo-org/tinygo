package sync_test

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestCondSignal tests waiting on a Cond and notifying it with Signal.
func TestCondSignal(t *testing.T) {
	// Create a Cond with a normal mutex.
	cond := sync.Cond{
		L: &sync.Mutex{},
	}
	cond.L.Lock()

	// Start a goroutine to signal us once we wait.
	var signaled uint32
	go func() {
		// Wait for the test goroutine to wait.
		cond.L.Lock()
		defer cond.L.Unlock()

		// Send a signal to the test goroutine.
		atomic.StoreUint32(&signaled, 1)
		cond.Signal()
	}()

	// Wait for a signal.
	// This will unlock the mutex, and allow the spawned goroutine to run.
	cond.Wait()
	if atomic.LoadUint32(&signaled) == 0 {
		t.Error("wait returned before a signal was sent")
	}
}

func TestCondBroadcast(t *testing.T) {
	// Create a Cond with an RWMutex.
	var mu sync.RWMutex
	cond := sync.Cond{
		L: mu.RLocker(),
	}

	// Start goroutines to wait for the broadcast.
	var wg sync.WaitGroup
	const n = 5
	for i := 0; i < n; i++ {
		wg.Add(1)
		mu.RLock()
		go func() {
			defer wg.Done()

			cond.Wait()
		}()
	}

	// Wait for all goroutines to start waiting.
	mu.Lock()

	// Broadcast to all of the waiting goroutines.
	cond.Broadcast()

	// Wait for all spawned goroutines to process the broadcast.
	mu.Unlock()
	wg.Wait()
}

// TestCondUnlockNotify verifies that a signal is processed even if it happens during the mutex unlock in Wait.
func TestCondUnlockNotify(t *testing.T) {
	// Create a Cond that signals itself when waiting.
	var cond sync.Cond
	cond.L = fakeLocker{cond.Signal}

	cond.Wait()
}

// fakeLocker is a fake sync.Locker where unlock calls an arbitrary function.
type fakeLocker struct {
	unlock func()
}

func (l fakeLocker) Lock()   {}
func (l fakeLocker) Unlock() { l.unlock() }
