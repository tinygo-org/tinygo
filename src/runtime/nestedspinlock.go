package runtime

import (
	"sync/atomic"
)

// NestedSpinLock is a NestedSpinLock implementation.
//
// A NestedSpinLock must not be copied after first use.
type NestedSpinLock struct {
	owner       uintptr
	nestedCount int
}

// Lock locks l.
// If the lock is already in use, the calling goroutine
// blocks until the locker is available.
func (l *NestedSpinLock) Lock() {
	tid := uintptr(currentMachineId())
	for !atomic.CompareAndSwapUintptr(&l.owner, 0, tid) {
		if atomic.LoadUintptr(&l.owner) == tid {
			break
		}
		// waiting for unlock
	}
	l.nestedCount++
}

// Unlock unlocks l.
func (l *NestedSpinLock) Unlock() {
	l.nestedCount--
	if l.nestedCount == 0 {
		atomic.StoreUintptr(&l.owner, 0)
	}
}
