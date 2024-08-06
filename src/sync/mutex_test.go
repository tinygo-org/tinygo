package sync_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func HammerMutex(m *sync.Mutex, loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		if i%3 == 0 {
			if m.TryLock() {
				m.Unlock()
			}
			continue
		}
		m.Lock()
		m.Unlock()
	}
	cdone <- true
}

func TestMutex(t *testing.T) {
	m := new(sync.Mutex)

	m.Lock()
	if m.TryLock() {
		t.Fatalf("TryLock succeeded with mutex locked")
	}
	m.Unlock()
	if !m.TryLock() {
		t.Fatalf("TryLock failed with mutex unlocked")
	}
	m.Unlock()

	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(m, 1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}

// TestMutexUncontended tests locking and unlocking a Mutex that is not shared with any other goroutines.
func TestMutexUncontended(t *testing.T) {
	var mu sync.Mutex

	// Lock and unlock the mutex a few times.
	for i := 0; i < 3; i++ {
		mu.Lock()
		mu.Unlock()
	}
}

// TestMutexConcurrent tests a mutex concurrently from multiple goroutines.
// It will fail if multiple goroutines hold the lock simultaneously.
func TestMutexConcurrent(t *testing.T) {
	var mu sync.Mutex
	var active uint
	var completed uint
	ok := true

	const n = 10
	for i := 0; i < n; i++ {
		j := i
		go func() {
			// Delay a bit.
			for k := j; k > 0; k-- {
				runtime.Gosched()
			}

			mu.Lock()

			// Increment the active counter.
			active++

			if active > 1 {
				// Multiple things are holding the lock at the same time.
				ok = false
			} else {
				// Delay a bit.
				for k := j; k < n; k++ {
					runtime.Gosched()
				}
			}

			// Decrement the active counter.
			active--

			// This is completed.
			completed++

			mu.Unlock()
		}()
	}

	// Wait for everything to finish.
	var done bool
	for !done {
		// Wait a bit for other things to run.
		runtime.Gosched()

		// Acquire the lock and check whether everything has completed.
		mu.Lock()
		done = completed == n
		mu.Unlock()
	}
	if !ok {
		t.Error("lock held concurrently")
	}
}

// TestRWMutexUncontended tests locking and unlocking an RWMutex that is not shared with any other goroutines.
func TestRWMutexUncontended(t *testing.T) {
	var mu sync.RWMutex

	// Lock the mutex exclusively and then unlock it.
	mu.Lock()
	mu.Unlock()

	// Acquire several read locks.
	const n = 5
	for i := 0; i < n; i++ {
		mu.RLock()
	}

	// Release all of the read locks.
	for i := 0; i < n; i++ {
		mu.RUnlock()
	}

	// Re-acquire the lock exclusively.
	mu.Lock()
	mu.Unlock()
}

// TestRWMutexWriteToRead tests the transition from a write lock to a read lock while contended.
func TestRWMutexWriteToRead(t *testing.T) {
	// Create a new RWMutex and acquire a write lock.
	var mu sync.RWMutex
	mu.Lock()

	const n = 3
	var readAcquires uint32
	var completed uint32
	var unlocked uint32
	var bad uint32
	for i := 0; i < n; i++ {
		go func() {
			// Acquire a read lock.
			mu.RLock()

			// Verify that the write lock is supposed to be released by now.
			if atomic.LoadUint32(&unlocked) == 0 {
				// The write lock is still being held.
				atomic.AddUint32(&bad, 1)
			}

			// Add ourselves to the read lock counter.
			atomic.AddUint32(&readAcquires, 1)

			// Wait for everything to hold the read lock simultaneously.
			for atomic.LoadUint32(&readAcquires) < n {
				runtime.Gosched()
			}

			// Notify of completion.
			atomic.AddUint32(&completed, 1)

			// Release the read lock.
			mu.RUnlock()
		}()
	}

	// Wait a bit for the goroutines to block.
	for i := 0; i < 3*n; i++ {
		runtime.Gosched()
	}

	// Release the write lock so that the goroutines acquire read locks.
	atomic.StoreUint32(&unlocked, 1)
	mu.Unlock()

	// Wait for everything to complete.
	for atomic.LoadUint32(&completed) < n {
		runtime.Gosched()
	}

	// Acquire another write lock.
	mu.Lock()

	if bad != 0 {
		t.Error("read lock acquired while write-locked")
	}
}

// TestRWMutexReadToWrite tests the transition from a read lock to a write lock while contended.
func TestRWMutexReadToWrite(t *testing.T) {
	// Create a new RWMutex and read-lock it several times.
	const n = 3
	var mu sync.RWMutex
	var readers uint32
	for i := 0; i < n; i++ {
		mu.RLock()
		readers++
	}

	// Start a goroutine to acquire a write lock.
	result := ^uint32(0)
	go func() {
		// Acquire a write lock.
		mu.Lock()

		// Check for active readers.
		readers := atomic.LoadUint32(&readers)

		mu.Unlock()

		// Report the number of active readers.
		atomic.StoreUint32(&result, readers)
	}()

	// Release the read locks.
	for i := 0; i < n; i++ {
		runtime.Gosched()
		atomic.AddUint32(&readers, ^uint32(0))
		mu.RUnlock()
	}

	// Wait for a result.
	var res uint32
	for res == ^uint32(0) {
		runtime.Gosched()
		res = atomic.LoadUint32(&result)
	}
	if res != 0 {
		t.Errorf("write lock acquired while %d readers were active", res)
	}
}
