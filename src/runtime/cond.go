//go:build !scheduler.none
// +build !scheduler.none

package runtime

import (
	"internal/task"
	"sync/atomic"
	"unsafe"
)

// notifiedPlaceholder is a placeholder task which is used to indicate that the condition variable has been notified.
var notifiedPlaceholder task.Task

// Cond is a simplified condition variable, useful for notifying goroutines of interrupts.
type Cond struct {
	t *task.Task
}

// Notify sends a notification.
// If the condition variable already has a pending notification, this returns false.
func (c *Cond) Notify() bool {
	for {
		t := (*task.Task)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t))))
		switch t {
		case nil:
			// Nothing is waiting yet.
			// Apply the notification placeholder.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t)), unsafe.Pointer(t), unsafe.Pointer(&notifiedPlaceholder)) {
				return true
			}
		case &notifiedPlaceholder:
			// The condition variable has already been notified.
			return false
		default:
			// Unblock the waiting task.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t)), unsafe.Pointer(t), nil) {
				runqueuePushBack(t)
				return true
			}
		}
	}
}

// Poll checks for a notification.
// If a notification is found, it is cleared and this returns true.
func (c *Cond) Poll() bool {
	for {
		t := (*task.Task)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t))))
		switch t {
		case nil:
			// No notifications are present.
			return false
		case &notifiedPlaceholder:
			// A notification arrived and there is no waiting goroutine.
			// Clear the notification and return.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t)), unsafe.Pointer(t), nil) {
				return true
			}
		default:
			// A task is blocked on the condition variable, which means it has not been notified.
			return false
		}
	}
}

// Wait for a notification.
// If the condition variable was previously notified, this returns immediately.
func (c *Cond) Wait() {
	cur := task.Current()
	for {
		t := (*task.Task)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t))))
		switch t {
		case nil:
			// Condition variable has not been notified.
			// Block the current task on the condition variable.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t)), nil, unsafe.Pointer(cur)) {
				task.Pause()
				return
			}
		case &notifiedPlaceholder:
			// A notification arrived and there is no waiting goroutine.
			// Clear the notification and return.
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.t)), unsafe.Pointer(t), nil) {
				return
			}
		default:
			panic("interrupt.Cond: condition variable in use by another goroutine")
		}
	}
}
