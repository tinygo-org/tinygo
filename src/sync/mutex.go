package sync

import (
	"internal/task"
	_ "unsafe"
)

// These mutexes assume there is only one thread of operation: no goroutines,
// interrupts or anything else.

type Mutex struct {
	locked  bool
	blocked task.Stack
}

//go:linkname scheduleTask runtime.runqueuePushBack
func scheduleTask(*task.Task)

func (m *Mutex) Lock() {
	if m.locked {
		// Push self onto stack of blocked tasks, and wait to be resumed.
		m.blocked.Push(task.Current())
		task.Pause()
		return
	}

	m.locked = true
}

func (m *Mutex) Unlock() {
	if !m.locked {
		panic("sync: unlock of unlocked Mutex")
	}

	// Wake up a blocked task, if applicable.
	if t := m.blocked.Pop(); t != nil {
		scheduleTask(t)
	} else {
		m.locked = false
	}
}

type RWMutex struct {
	m       Mutex
	readers uint32
}

func (rw *RWMutex) Lock() {
	rw.m.Lock()
}

func (rw *RWMutex) Unlock() {
	rw.m.Unlock()
}

func (rw *RWMutex) RLock() {
	if rw.readers == 0 {
		rw.m.Lock()
	}
	rw.readers++
}

func (rw *RWMutex) RUnlock() {
	if rw.readers == 0 {
		panic("sync: unlock of unlocked RWMutex")
	}
	rw.readers--
	if rw.readers == 0 {
		rw.m.Unlock()
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
