package task

// Barebones semaphore implementation.
// The main limitation is that if there are multiple waiters, a single Post()
// call won't do anything. Only when Post() has been called to awaken all
// waiters will the waiters proceed.
// This limitation is not a problem when there will only be a single waiter.
type Semaphore struct {
	futex Futex
}

// Post (unlock) the semaphore, incrementing the value in the semaphore.
func (s *Semaphore) Post() {
	newValue := s.futex.Add(1)
	if newValue == 0 {
		s.futex.WakeAll()
	}
}

// Wait (lock) the semaphore, decrementing the value in the semaphore.
func (s *Semaphore) Wait() {
	delta := int32(-1)
	value := s.futex.Add(uint32(delta))
	for {
		if int32(value) >= 0 {
			// Semaphore unlocked!
			return
		}
		s.futex.Wait(value)
		value = s.futex.Load()
	}
}
