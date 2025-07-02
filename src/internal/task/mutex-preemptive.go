//go:build !tinygo.unicore

package task

// Futex-based mutex.
// This is largely based on the paper "Futexes are Tricky" by Ulrich Drepper.
// It describes a few ways to implement mutexes using a futex, and how some
// seemingly-obvious implementations don't exactly work as intended.
// Unfortunately, Go atomic operations work slightly differently so we can't
// copy the algorithm verbatim.
//
// The implementation works like this. The futex can have 3 different values,
// depending on the state:
//
//   - 0: the futex is currently unlocked.
//   - 1: the futex is locked, but is uncontended. There is one special case: if
//     a contended futex is unlocked, it is set to 0. It is possible for another
//     thread to lock the futex before the next waiter is woken. But because a
//     waiter will be woken (if there is one), it will always change to 2
//     regardless. So this is not a problem.
//   - 2: the futex is locked, and is contended. At least one thread is trying
//     to obtain the lock (and is in the contended loop, see below).
//
// For the paper, see:
// https://dept-info.labri.fr/~denis/Enseignement/2008-IR/Articles/01-futex.pdf)

type Mutex struct {
	futex Futex
}

func (m *Mutex) Lock() {
	// Fast path: try to take an uncontended lock.
	if m.futex.CompareAndSwap(0, 1) {
		// We obtained the mutex.
		return
	}

	// The futex is contended, so we enter the contended loop.
	// If we manage to change the futex from 0 to 2, we managed to take the
	// lock. Else, we have to wait until a call to Unlock unlocks this mutex.
	// (Unlock will wake one waiter when it finds the futex is set to 2 when
	// unlocking).
	for m.futex.Swap(2) != 0 {
		// Wait until we get resumed in Unlock.
		m.futex.Wait(2)
	}
}

func (m *Mutex) Unlock() {
	if old := m.futex.Swap(0); old == 0 {
		// Mutex wasn't locked before.
		panic("sync: unlock of unlocked Mutex")
	} else if old == 2 {
		// Mutex was a contended lock, so we need to wake the next waiter.
		m.futex.Wake()
	}
}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (m *Mutex) TryLock() bool {
	// Fast path: try to take an uncontended lock.
	if m.futex.CompareAndSwap(0, 1) {
		// We obtained the mutex.
		return true
	}
	return false
}
