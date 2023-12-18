package spinlock

import (
	"sync/atomic"
)

// SpinLock is a spinlock implementation.
//
// A SpinLock must not be copied after first use.
type SpinLock struct {
	lock uintptr
}

// Lock locks l.
// If the lock is already in use, the calling goroutine
// blocks until the locker is available.
func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) {
		// waiting for unlock
	}
}

// Unlock unlocks l.
func (l *SpinLock) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}
