package runtime

// This file implements the internal/poll semaphore linknames
// (runtime_Semacquire / runtime_Semrelease) for upstream internal/poll
// (fd_mutex.go, fd_unix.go), which calls them to block/unblock fd reference
// counting and Close waits.
//
// Upstream Go keys its semaphore waiters by the semaphore address in a
// runtime-side table (semaRoot queue in runtime/sema.go) and tracks the
// waiter count separately (nwait). We do the same here: a runtime-side map
// from the semaphore word address to a waiter stack, guarded by lockAtomics.
// The semaphore word itself is only a permit counter and is never
// reinterpreted as another struct, so the cooperative task.Futex layout
// (Uint32 + waiters Stack) cannot overlay adjacent fields of the caller
// (fdMutex.rsema/wsema, FD.csema).

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

// semaWaiter is one task blocked in semacquire on a semaphore word.
type semaWaiter struct {
	next *semaWaiter
	task *task.Task
}

// semaWaiters maps a semaphore word address to its LIFO waiter stack.
// Guarded by lockAtomics (interrupts off on cooperative, a real lock on
// preemptive schedulers), so Push/Pop are atomic w.r.t. semrelease.
var semaWaiters = map[uintptr]*semaWaiter{}

//go:linkname semacquire internal/poll.runtime_Semacquire
func semacquire(sema *uint32) {
	for {
		v := atomic.LoadUint32(sema)
		if v != 0 && atomic.CompareAndSwapUint32(sema, v, v-1) {
			return
		}
		// The word is 0: register as a waiter, then re-check the word to
		// avoid a lost wakeup (semrelease increments under lockAtomics and
		// pops a waiter in the same critical section).
		w := &semaWaiter{task: task.Current()}
		mask := lockAtomics()
		w.next = semaWaiters[uintptr(unsafe.Pointer(sema))]
		semaWaiters[uintptr(unsafe.Pointer(sema))] = w
		v = atomic.LoadUint32(sema)
		if v != 0 {
			// Raced with a release: take the permit instead of sleeping.
			semaWaiters[uintptr(unsafe.Pointer(sema))] = w.next
			unlockAtomics(mask)
			if atomic.CompareAndSwapUint32(sema, v, v-1) {
				return
			}
			// CAS failed; retry the whole loop.
			continue
		}
		unlockAtomics(mask)
		task.Pause()
		// Awoken by semrelease: the permit is ours (semrelease does not
		// increment for a popped waiter), so just return.
		return
	}
}

//go:linkname semrelease internal/poll.runtime_Semrelease
func semrelease(sema *uint32) {
	// The whole operation runs under lockAtomics: the waiter check, the pop,
	// and the increment are one critical section, so a concurrent semacquire
	// can never push a waiter after we checked and before we increment
	// (that interleaving would lose the wakeup under -scheduler=threads).
	mask := lockAtomics()
	key := uintptr(unsafe.Pointer(sema))
	if w := semaWaiters[key]; w != nil {
		// A waiter is asleep: hand the permit directly to it (do not
		// increment the word, so the waiter's return path needs no CAS).
		semaWaiters[key] = w.next
		unlockAtomics(mask)
		scheduleTask(w.task)
		return
	}
	atomic.AddUint32(sema, 1)
	unlockAtomics(mask)
}
