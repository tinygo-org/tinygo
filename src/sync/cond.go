package sync

import (
	"internal/task"
	"unsafe"
)

// Condition variable.
// A goroutine that called Wait() can be in one of a few states depending on the
// Task.Data field:
// - When entering Wait, and before going to sleep, the data field is 0.
// - When the goroutine that calls Wait changes its data value from 0 to 1, it
//   is going to sleep. It has not been awoken early.
// - When instead a call to Signal or Broadcast can change the data field from 0
//   to 1, it will _not_ go to sleep but be signalled early.
//   This can happen when a concurrent call to Signal happens, or the Unlock
//   function calls Signal for some reason.

type Cond struct {
	L Locker

	blocked task.Stack
	lock    task.PMutex
}

func NewCond(l Locker) *Cond {
	return &Cond{L: l}
}

func (c *Cond) trySignal() bool {
	// Pop a blocked task off of the stack, and schedule it if applicable.
	t := c.blocked.Pop()
	if t != nil {
		if t.SynctestBubble != nil && task.Current().SynctestBubble != t.SynctestBubble {
			runtimeFatal("semaphore wake of synctest goroutine from outside bubble")
		}
		dataPtr := (*task.Uint32)(unsafe.Pointer(&t.Data))

		// The data value is 0 when the task is not yet sleeping, and 1 when it is.
		if dataPtr.Swap(1) != 0 {
			// The value was already 1, so the task went to sleep (or is about to go
			// to sleep). Schedule the task to be resumed.
			scheduleTask(t)
		}
		return true
	}

	// There was nothing to signal.
	return false
}

func (c *Cond) Signal() {
	c.lock.Lock()
	c.trySignal()
	c.lock.Unlock()
}

func (c *Cond) Broadcast() {
	// Signal everything.
	c.lock.Lock()
	for c.trySignal() {
	}
	c.lock.Unlock()
}

func (c *Cond) Wait() {
	// Mark us as not yet signalled or sleeping.
	t := task.Current()
	dataPtr := (*task.Uint32)(unsafe.Pointer(&t.Data))
	dataPtr.Store(0)

	// Add us to the list of waiting goroutines.
	c.lock.Lock()
	c.blocked.Push(t)
	c.lock.Unlock()

	transition := synctestBlockBegin(t)

	// Temporarily unlock L.
	c.L.Unlock()

	// Re-acquire the lock before returning.
	defer c.L.Lock()

	// If we were signaled while unlocking, immediately complete.
	if dataPtr.Swap(1) != 0 {
		if transition {
			synctestBlockEnd(t, false)
		}
		// The data value was already 1, so we got a signal already (and weren't
		// scheduled because trySignal was the first to change the value).
		return
	}

	// We were the first to change the value from 0 to 1, meaning we did not get
	// a signal during the call to Unlock(). So we wait until we do get a
	// signal.
	if transition {
		synctestBlockEnd(t, true)
	}
	task.Pause()
}

//go:linkname scheduleTask runtime.scheduleTask
func scheduleTask(*task.Task)

//go:linkname synctestBlockBegin runtime.synctestBlockBegin
func synctestBlockBegin(*task.Task) bool

//go:linkname synctestBlockEnd runtime.synctestBlockEnd
func synctestBlockEnd(*task.Task, bool)
