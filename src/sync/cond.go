package sync

type Cond struct {
	L Locker

	ch chan struct{}
}

func (c *Cond) Wait() {
	c.L.Unlock()

	if c.ch == nil {
		// first waiter, initialize channel
		c.ch = make(chan struct{})
	}

	<-c.ch

	c.L.Lock()
}

// trySignal attempts to signal a waiting goroutine, and returns true if successful.
func (c *Cond) trySignal() bool {
	select {
	case c.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (c *Cond) Signal() {
	c.trySignal()
}

func (c *Cond) Broadcast() {
	for c.trySignal() {
	}
}
