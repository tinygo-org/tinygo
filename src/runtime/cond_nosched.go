//go:build scheduler.none

package runtime

import "runtime/interrupt"

// Cond is a simplified condition variable, useful for notifying goroutines of interrupts.
type Cond struct {
	notified bool
}

// Notify sends a notification.
// If the condition variable already has a pending notification, this returns false.
func (c *Cond) Notify() bool {
	i := interrupt.Disable()
	prev := c.notified
	c.notified = true
	interrupt.Restore(i)
	return !prev
}

// Poll checks for a notification.
// If a notification is found, it is cleared and this returns true.
func (c *Cond) Poll() bool {
	i := interrupt.Disable()
	notified := c.notified
	c.notified = false
	interrupt.Restore(i)
	return notified
}

// Wait for a notification.
// If the condition variable was previously notified, this returns immediately.
func (c *Cond) Wait() {
	for !c.Poll() {
		waitForEvents()
	}
}
