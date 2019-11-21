package sync

// These mutexes assume there is only one thread of operation: no goroutines,
// interrupts or anything else.

type Locker interface {
	Lock()
	Unlock()
}

type Mutex struct {
	ch chan struct{}
}

func (m *Mutex) Lock() {
	if m.ch == nil {
		m.ch = make(chan struct{}, 1)
	}
	m.ch <- struct{}{}
}

func (m *Mutex) Unlock() {
	select {
	case <-m.ch:
	default:
		panic("sync: unlock of unlocked Mutex")
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
