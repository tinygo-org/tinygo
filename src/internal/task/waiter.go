package task

import (
	"runtime/interrupt"
	"unsafe"
)

// A waiter waits for an interrupt to fire. Call Wait() from a goroutine to
// pause it, and call Resume() from an interrupt handler to resume the
// goroutine.
type Waiter struct {
	waiting *Task
	lock    PMutex
}

var resumeTaskSentinel = (*Task)(unsafe.Pointer(uintptr(2)))
var workingTaskSentinel = (*Task)(unsafe.Pointer(uintptr(1)))

// Wait from a main goroutine until an interrupt calls Resume(). It will also
// return if the interrupt called Resume() before the Wait() call (to avoid a
// race condition).
func (w *Waiter) Wait() {
	if interrupt.In() {
		runtimePanic("Waiter: called Wait from interrupt")
	}

	mask := interrupt.Disable()
	w.lock.Lock()
	switch w.waiting {
	case nil, workingTaskSentinel:
		w.waiting = Current()
		w.lock.Unlock()
		interrupt.Restore(mask)
		Pause()
	case resumeTaskSentinel:
		// Marked as 'resume now', can return immediately.
		w.waiting = workingTaskSentinel
		w.lock.Unlock()
		interrupt.Restore(mask)
	default:
		w.lock.Unlock()
		interrupt.Restore(mask)
		runtimePanic("Waiter: task is waiting already")
	}
}

// Resume a waiting goroutine from an interrupt handler. If there is a goroutine
// waiting, it will resume. If not, the next one that calls Wait() will
// immediately resume.
func (w *Waiter) Resume() {
	if !interrupt.In() {
		runtimePanic("Waiter: called Resume outside interrupt")
	}

	waiting := w.waiting
	switch waiting {
	case nil, workingTaskSentinel:
		w.waiting = resumeTaskSentinel
	case resumeTaskSentinel:
		// Called Resume() twice in a row, this may indicate a bug at the caller
		// site.
	default:
		// Schedule the given task to resume.
		// If it's not yet paused, it will immediately resume on the next call
		// to Pause().
		w.waiting = workingTaskSentinel
		scheduleTask(waiting)
	}
}

// Return true if Done has not been called after a Resume call.
// This is typically used to indicate the task outside the interrupt is still
// being worked on.
func (w *Waiter) Working() bool {
	switch w.waiting {
	case workingTaskSentinel, resumeTaskSentinel:
		return true
	default: // nil, <*task.Task>
		return false
	}
}

// Done can be called outside interrupt context to indicate the task this waiter
// was waiting for is done, and Working() can return false. It is entirely
// optional, if not called Working will continue to return true until it is
// blocked in Wait() but the waiter will function normally otherwise.
func (w *Waiter) Done() {
	if interrupt.In() {
		runtimePanic("Waiter: called Done from interrupt")
	}

	mask := interrupt.Disable()
	w.lock.Lock()
	if w.waiting == workingTaskSentinel {
		w.waiting = nil
	}
	w.lock.Unlock()
	interrupt.Restore(mask)
}
