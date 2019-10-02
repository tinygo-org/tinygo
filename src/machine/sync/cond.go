package sync

import "unsafe"

//go:linkname pushInt runtime.pushInt
func pushInt(t unsafe.Pointer)

//go:linkname blockTask runtime.blockTask
func blockTask(t unsafe.Pointer, chain *unsafe.Pointer)

//go:linkname getCoroutine runtime.getCoroutine
func getCoroutine() unsafe.Pointer

//go:linkname yield runtime.yield
func yield()

//go:linkname unblock runtime.unblock
func unblock(t unsafe.Pointer) unsafe.Pointer

// Cond is a condition variable that can be used by interrupts to notify goroutines.
type Cond struct {
	fired, fireAck bool
	t unsafe.Pointer
	chain unsafe.Pointer
}

// Notify marks the Cond as completed, and unblocks all blockers.
// An interruptor of this call must not try to notify.
// Returns true if and only if the condition had not previously been notified.
func (c *Cond) Notify() bool {
	if c.fired {
		// already fired once - do nothing
		return false
	}

	// set flag so that no additional tasks block on this cond
	c.fired = true

	if c.t != nil {
		// there is a task blocked on this Cond, push it into the wakeup queue
		pushInt(c.t)
		c.fireAck = true
	}

	// we think we did something
	return true
}

// Wait blocks until the Cond is notified.
// If the Cond is notified before this call, this will unblock immediately.
// This does not clear the notification, so subsequent calls to Wait will not block.
func (c *Cond) Wait() {
	if c.t != nil {
		// task already blocked on Cond
		// enqueue ourselves after it
		blockTask(getCoroutine(), &c.chain)

		// wait for the first blocker to wake us up
		yield()
		return
	}

	// place ourselves as a blocker in the Cond
	// we need to do this before checking c.fired, in case an interrupt fires here
	c.t = getCoroutine()

	switch {
	case c.fireAck:
		// an interrupt called notify and saw our coroutine
		// we are on the wakeup queue
		// we need to yield, and will be immediately awoken
		yield()
	case c.fired:
		// an interrupt called notify but did not see our coroutine
		// we are not on the wakeup queue
		// we do not need to wait
		c.t = nil
	default:
		// nothing has been notified yet
		// wait for notification
		yield()
	}

	// detatch ourself
	c.t = nil

	// unblock all other things that are blocked on the cond
	for t := c.chain; t != nil; t = unblock(t) {}
	c.chain = nil

	// acknowledge ack
	c.fireAck = false
}

// Clear resets the condition variable.
// Subsequent calls to Wait will block until Notify is called again.
func (c *Cond) Clear() {
	if c.t != nil {
		panic("cannot clear a blocked condition variable")
	}
	c.fired, c.fireAck = false, false
}
