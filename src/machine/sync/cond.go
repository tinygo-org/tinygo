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
	state condState
	t unsafe.Pointer
	chain unsafe.Pointer
}

type condState uint8

const (
	condStateReady condState = iota
	condStateFired
	condStateFiredAck
)

// Notify marks the Cond as completed, and unblocks all blockers.
// An interruptor of this call must not try to notify.
// Returns true if and only if the condition had not previously been notified.
func (c *Cond) Notify() bool {
	switch c.state {
	case condStateReady:
		// condition variable has not been previously notified
		if c.t != nil {
			// there is a task blocked on this Cond, push it into the wakeup queue and acknowledge that we noticed it
			pushInt(c.t)
			c.state = condStateFiredAck
		} else {
			// the Cond has not yet been blocked on, mark it as notified
			c.state = condStateFired
		}

		// we think we did something
		return true
	default:
		// already fired, we did nothing
		return false
	}
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

	switch c.state {
	case condStateFiredAck:
		// an interrupt called notify and saw our coroutine
		// we are on the wakeup queue
		// we need to yield, and will be immediately awoken
		yield()
	case condStateFired:
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

	// finalize state
	c.state = condStateFired
}

// Clear resets the condition variable.
// Subsequent calls to Wait will block until Notify is called again.
func (c *Cond) Clear() {
	if c.t != nil {
		panic("cannot clear a blocked condition variable")
	}
	c.state = condStateReady
}
