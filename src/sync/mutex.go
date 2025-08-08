package sync

import (
	"internal/task"
)

type Mutex = task.Mutex

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(msg string)

type RWMutex struct {
	// Reader count, with the number of readers that currently have read-locked
	// this mutex.
	// The value can be in two states: one where 0 means no readers and another
	// where -rwMutexMaxReaders means no readers. A base of 0 is normal
	// uncontended operation, a base of -rwMutexMaxReaders means a writer has
	// the lock or is trying to get the lock. In the second case, readers should
	// wait until the reader count becomes non-negative again to give the writer
	// a chance to obtain the lock.
	readers task.Futex

	// Writer futex, normally 0. If there is a writer waiting until all readers
	// have unlocked, this value is 1. It will be changed to a 2 (and get a
	// wake) when the last reader unlocks.
	writer task.Futex

	// Writer lock. Held between Lock() and Unlock().
	writerLock Mutex
}

const rwMutexMaxReaders = 1 << 30

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
func (rw *RWMutex) Lock() {
	// Exclusive lock for writers.
	rw.writerLock.Lock()

	// Flag that we need to be awakened after the last read-lock unlocks.
	rw.writer.Store(1)

	// Signal to readers that they can't lock this mutex anymore.
	n := uint32(rwMutexMaxReaders)
	waiting := rw.readers.Add(-n)
	if int32(waiting) == -rwMutexMaxReaders {
		// All readers were already unlocked, so we don't need to wait for them.
		rw.writer.Store(0)
		return
	}

	// There is at least one reader.
	// Wait until all readers are unlocked. The last reader to unlock will set
	// rw.writer to 2 and awaken us.
	for rw.writer.Load() == 1 {
		rw.writer.Wait(1)
	}
	rw.writer.Store(0)
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked [RWMutex] is not associated with a particular
// goroutine. One goroutine may [RWMutex.RLock] ([RWMutex.Lock]) a RWMutex and then
// arrange for another goroutine to [RWMutex.RUnlock] ([RWMutex.Unlock]) it.
func (rw *RWMutex) Unlock() {
	// Signal that new readers can lock this mutex.
	waiting := rw.readers.Add(rwMutexMaxReaders)
	if waiting != 0 {
		// Awaken all waiting readers.
		rw.readers.WakeAll()
	}

	// Done with this lock (next writer can try to get a lock).
	rw.writerLock.Unlock()
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (rw *RWMutex) TryLock() bool {
	// Check for active writers
	if !rw.writerLock.TryLock() {
		return false
	}
	// Have write lock, now check for active readers
	n := uint32(rwMutexMaxReaders)
	if !rw.readers.CompareAndSwap(0, -n) {
		// Active readers, give up write lock
		rw.writerLock.Unlock()
		return false
	}
	return true
}

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the [RWMutex] type.
func (rw *RWMutex) RLock() {
	// Add us as a reader.
	newVal := rw.readers.Add(1)

	// Wait until the RWMutex is available for readers.
	for int32(newVal) <= 0 {
		rw.readers.Wait(newVal)
		newVal = rw.readers.Load()
	}
}

// RUnlock undoes a single [RWMutex.RLock] call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
func (rw *RWMutex) RUnlock() {
	// Remove us as a reader.
	one := uint32(1)
	readers := int32(rw.readers.Add(-one))

	// Check whether RUnlock was called too often.
	if readers == -1 || readers == (-rwMutexMaxReaders)-1 {
		runtimePanic("sync: RUnlock of unlocked RWMutex")
	}

	if readers == -rwMutexMaxReaders {
		// This was the last read lock. Check whether we need to wake up a write
		// lock.
		if rw.writer.CompareAndSwap(1, 2) {
			rw.writer.Wake()
		}
	}
}

// TryRLock tries to lock rw for reading and reports whether it succeeded.
//
// Note that while correct uses of TryRLock do exist, they are rare,
// and use of TryRLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (rw *RWMutex) TryRLock() bool {
	for {
		c := rw.readers.Load()
		if c < 0 {
			// There is a writer waiting or writing.
			return false
		}
		if rw.readers.CompareAndSwap(c, c+1) {
			// Read lock obtained.
			return true
		}
	}
}

type Locker interface {
	Lock()
	Unlock()
}

// RLocker returns a Locker interface that implements
// the Lock and Unlock methods by calling rw.RLock and rw.RUnlock.
func (rw *RWMutex) RLocker() Locker {
	return (*rlocker)(rw)
}

type rlocker RWMutex

func (r *rlocker) Lock()   { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }
