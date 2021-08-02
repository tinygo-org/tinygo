package sync

import "internal/task"

type Cond struct {
	L Locker

	unlocking *earlySignal
	blocked   task.Stack
}

// earlySignal is a type used to implement a stack for signalling waiters while they are unlocking.
type earlySignal struct {
	next *earlySignal

	signaled bool
}

func NewCond(l Locker) *Cond {
	return &Cond{L: l}
}

func (c *Cond) trySignal() bool {
	// Pop a blocked task off of the stack, and schedule it if applicable.
	t := c.blocked.Pop()
	if t != nil {
		scheduleTask(t)
		return true
	}

	// If there any tasks which are currently unlocking, signal one.
	if c.unlocking != nil {
		c.unlocking.signaled = true
		c.unlocking = c.unlocking.next
		return true
	}

	// There was nothing to signal.
	return false
}

func (c *Cond) Signal() {
	c.trySignal()
}

func (c *Cond) Broadcast() {
	// Signal everything.
	for c.trySignal() {
	}
}

func (c *Cond) Wait() {
	// Add an earlySignal frame to the stack so we can be signalled while unlocking.
	early := earlySignal{
		next: c.unlocking,
	}
	c.unlocking = &early

	// Temporarily unlock L.
	c.L.Unlock()

	// Re-acquire the lock before returning.
	defer c.L.Lock()

	// If we were signaled while unlocking, immediately complete.
	if early.signaled {
		return
	}

	// Remove the earlySignal frame.
	prev := c.unlocking
	for prev != nil && prev.next != &early {
		prev = prev.next
	}
	if prev != nil {
		prev.next = early.next
	} else {
		c.unlocking = early.next
	}

	// Wait for a signal.
	c.blocked.Push(task.Current())
	task.Pause()
}
