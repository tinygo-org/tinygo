package sync_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mutex interface {
	Lock()
	Unlock()
	TryLock() bool
}

func HammerMutex(m mutex, loops int, cdone chan bool) {
	for i := range loops {
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
	for range 10 {
		go HammerMutex(m, 1000, c)
	}
	for range 10 {
		<-c
	}
}

// TestMutexUncontended tests locking and unlocking a Mutex that is not shared with any other goroutines.
func TestMutexUncontended(t *testing.T) {
	var mu sync.Mutex

	// Lock and unlock the mutex a few times.
	for range 3 {
		mu.Lock()
		mu.Unlock()
	}
}

// TestMutexConcurrent tests a mutex concurrently from multiple goroutines.
// It will fail if multiple goroutines hold the lock simultaneously.
func TestMutexConcurrent(t *testing.T) {
	var mu sync.Mutex
	var active atomic.Uint32
	var completed atomic.Uint32
	var fail atomic.Uint32

	const n = 10
	for i := range n {
		j := i
		go func() {
			// Delay a bit.
			for k := j; k > 0; k-- {
				runtime.Gosched()
			}

			mu.Lock()

			// Increment the active counter.
			nowActive := active.Add(1)

			if nowActive > 1 {
				// Multiple things are holding the lock at the same time.
				fail.Store(1)
			} else {
				// Delay a bit.
				for k := j; k < n; k++ {
					runtime.Gosched()
				}
			}

			// Decrement the active counter.
			var one = 1
			active.Add(uint32(-one))

			// This is completed.
			completed.Add(1)

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
		done = completed.Load() == n
		mu.Unlock()
	}
	if fail.Load() != 0 {
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
	for range n {
		mu.RLock()
	}

	// Release all of the read locks.
	for range n {
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
	var readAcquires atomic.Uint32
	var completed atomic.Uint32
	var unlocked atomic.Uint32
	var bad uint32
	for range n {
		go func() {
			// Acquire a read lock.
			mu.RLock()

			// Verify that the write lock is supposed to be released by now.
			if unlocked.Load() == 0 {
				// The write lock is still being held.
				atomic.AddUint32(&bad, 1)
			}

			// Add ourselves to the read lock counter.
			readAcquires.Add(1)

			// Wait for everything to hold the read lock simultaneously.
			for readAcquires.Load() < n {
				runtime.Gosched()
			}

			// Notify of completion.
			completed.Add(1)

			// Release the read lock.
			mu.RUnlock()
		}()
	}

	// Wait a bit for the goroutines to block.
	for range 3 * n {
		runtime.Gosched()
	}

	// Release the write lock so that the goroutines acquire read locks.
	unlocked.Store(1)
	mu.Unlock()

	// Wait for everything to complete.
	for completed.Load() < n {
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
	for range n {
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
	for range n {
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

func TestRWMutex(t *testing.T) {
	m := new(sync.RWMutex)

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
	for range 10 {
		go HammerMutex(m, 1000, c)
	}
	for range 10 {
		<-c
	}
}

// A writer must not wait forever for readers that arrive after it does.
func TestRWMutexWriterNotStarvedByLateReaders(t *testing.T) {
	var m sync.RWMutex
	locked := make(chan struct{})

	// A reader holds the lock, so the writer has to wait for it.
	m.RLock()

	go func() {
		m.Lock()
		m.Unlock()
		close(locked)
	}()
	// Give the writer time to register itself.
	time.Sleep(50 * time.Millisecond)

	// A second reader arrives while the writer waits. It must queue behind the
	// writer and not join the count that the writer waits on.
	secondReader := make(chan struct{})
	go func() {
		m.RLock()
		m.RUnlock()
		close(secondReader)
	}()
	time.Sleep(50 * time.Millisecond)

	// The first reader leaves, which is the last reader the writer waits for.
	m.RUnlock()

	select {
	case <-locked:
	case <-time.After(10 * time.Second):
		t.Fatal("the writer did not wake after the last reader unlocked")
	}
	select {
	case <-secondReader:
	case <-time.After(10 * time.Second):
		t.Fatal("the queued reader did not wake after the writer unlocked")
	}
}

// Readers released by an Unlock must take a permit and not read the reader
// count again, which a writer that arrives in between can change back.
func TestRWMutexHandoffToQueuedReaders(t *testing.T) {
	var m sync.RWMutex
	const iterations = 5000
	done := make(chan struct{})

	for range 4 {
		go func() {
			for range iterations {
				m.RLock()
				m.RUnlock()
			}
			done <- struct{}{}
		}()
	}
	for range 3 {
		go func() {
			for range iterations {
				m.Lock()
				m.Unlock()
			}
			done <- struct{}{}
		}()
	}

	for range 7 {
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Fatal("readers and writers stopped while they hand the lock over")
		}
	}
}
