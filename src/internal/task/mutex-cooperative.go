//go:build tinygo.unicore

package task

type Mutex struct {
	locked  bool
	blocked Stack
}

func (m *Mutex) Lock() {
	if m.locked {
		// Push self onto stack of blocked tasks, and wait to be resumed.
		m.blocked.Push(Current())
		Pause()
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

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that while correct uses of TryLock do exist, they are rare,
// and use of TryLock is often a sign of a deeper problem
// in a particular use of mutexes.
func (m *Mutex) TryLock() bool {
	if m.locked {
		return false
	}
	m.Lock()
	return true
}
