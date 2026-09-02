package sync

import (
	"internal/task"
	"unsafe"
)

// Cond is a condition variable.
type Cond struct {
	L Locker

	blocked task.Stack
	lock    task.PMutex
}

// A waiting task stores one of these states in Task.Data.
// Signal schedules only a task that has reached condBlocked.
const (
	condWaiting = iota
	condCommitting
	condSignaled
	condBlocked
)

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

		if dataPtr.Swap(condSignaled) == condBlocked {
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
	dataPtr.Store(condWaiting)

	// Add us to the list of waiting goroutines.
	c.lock.Lock()
	c.blocked.Push(t)
	c.lock.Unlock()

	transition := synctestBlockBegin(t)

	// Temporarily unlock L.
	c.L.Unlock()

	// Re-acquire the lock before returning.
	defer c.L.Lock()

	// Commit to blocking unless a signal arrived while unlocking.
	if !dataPtr.CompareAndSwap(condWaiting, condCommitting) {
		if transition {
			synctestBlockEnd(t, false)
		}
		return
	}

	if transition {
		if !synctestBlockCommit(t, dataPtr, condCommitting, condBlocked) {
			return
		}
	} else if !dataPtr.CompareAndSwap(condCommitting, condBlocked) {
		return
	}
	task.Pause()
}

//go:linkname scheduleTask runtime.scheduleTask
func scheduleTask(*task.Task)

//go:linkname synctestBlockBegin runtime.synctestBlockBegin
func synctestBlockBegin(*task.Task) bool

//go:linkname synctestBlockEnd runtime.synctestBlockEnd
func synctestBlockEnd(*task.Task, bool)

//go:linkname synctestBlockCommit runtime.synctestBlockCommit
func synctestBlockCommit(*task.Task, *task.Uint32, uint32, uint32) bool
