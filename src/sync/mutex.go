package sync

// These mutexes assume there is only one thread of operation: no goroutines,
// interrupts or anything else.

type Mutex struct {
	locked bool
}

func (m *Mutex) Lock() {
	if m.locked {
		panic("todo: block on locked mutex")
	}
	m.locked = true
}

func (m *Mutex) Unlock() {
	if !m.locked {
		panic("sync: unlock of unlocked Mutex")
	}
	m.locked = false
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
