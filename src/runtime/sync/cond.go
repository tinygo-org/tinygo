package sync

import (
	"internal/task"
	"runtime/volatile"
	"unsafe"
)

//go:linkname pushInterrupt runtime.pushInterrupt
func pushInterrupt(*task.Task)

// IntCond is a condition variable which can be safely signaled from an interrupt.
type IntCond struct {
	// setup is a flag which is used to indicate that a waiting goroutine is currently being set up.
	// While this set, the signaler will not push the next task onto the runqueue.
	setup volatile.Register8

	// signaledEarly is a flag used to indicate that the condition variable was signaled during or before setup.
	signaledEarly volatile.Register8

	// signaledNext is a flag used to indicate that the task stored in next has been signaled.
	signaledNext volatile.Register8

	// next is the next waiting task.
	next *task.Task
}

// Wait for a signal.
// This cannot be used outside of a goroutine.
// This is not safe for concurrent use.
func (c *IntCond) Wait() {
	if c.next != nil {
		panic("concurrent waiters on interrupt condition")
	}

	// Store the task pointer to be reawakened.
	c.setup.Set(1)
	volatile.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.next)), unsafe.Pointer(task.Current()))
	c.setup.Set(0)
	if c.signaledEarly.Get() != 0 {
		// We were signaled during setup or before.
		volatile.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.next)), nil)
		c.signaledEarly.Set(0)
		return
	}

	// Wait to be reawkakened.
	task.Pause()

	// Remove the task pointer.
	volatile.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&c.next)), nil)

	// Clear the signal flag.
	c.signaledNext.Set(0)
}

// Notify a waiting goroutine.
// If no goroutine is waiting, this flags the condition variable as notified and the next goroutine to wait will wake up immediately.
// If the condition variable already has a pending signal, this does nothing and returns false.
// Notify cannot be safely nested.
func (c *IntCond) Notify() bool {
	switch {
	case c.signaledEarly.Get() != 0 || c.signaledNext.Get() != 0:
		// There is already an unconsumed signal.
		return false
	case c.setup.Get() != 0 || c.next == nil:
		// Nothing is waiting yet - propogate an early signal.
		c.signaledEarly.Set(1)
	default:
		// Reawaken the waiting task.
		c.signaledNext.Set(1)
		pushInterrupt(c.next)
	}

	return true
}
