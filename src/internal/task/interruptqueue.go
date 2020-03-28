package task

import "runtime/volatile"

// InterruptQueue is a specialized version of Queue, designed for working with interrupts.
// It can be safely pushed to from an interrupt (assuming that a memory reference to the task remains elsewhere), and popped outside of an interrupt.
// It cannot be pushed to outside of an interrupt or popped inside an interrupt.
type InterruptQueue struct {
	// This implementation uses a double-buffer of queues.
	// bufSelect contains the index of the queue currently available for pop operations.
	// The opposite queue is available for push operations.
	bufSelect volatile.Register8
	queues    [2]Queue
}

// Push a task onto the queue.
// This can only be safely called from inside an interrupt.
func (q *InterruptQueue) Push(t *Task) {
	// Avoid nesting interrupts inside here.
	var nest nonest
	nest.Lock()
	defer nest.Unlock()

	// Push to inactive queue.
	q.queues[1-q.bufSelect.Get()].Push(t)
}

// Check if the queue is empty.
// This will return false if any tasks were pushed strictly before this call.
// If any pushes occur during the call, the queue may or may not be marked as empty.
// This cannot be safely called inside an interrupt.
func (q *InterruptQueue) Empty() bool {
	// Check currently active queue.
	active := q.bufSelect.Get() & 1
	if !q.queues[active].Empty() {
		return false
	}

	// Swap to other queue.
	active ^= 1
	q.bufSelect.Set(active)

	// Check other queue.
	return q.queues[active].Empty()
}

// Pop removes a single task from the queue.
// This will return nil if the queue is empty (with the same semantics as Empty).
// This cannot be safely called inside an interrupt.
func (q *InterruptQueue) Pop() *Task {
	// Select non-empty queue if one exists.
	if q.Empty() {
		return nil
	}

	// Pop from active queue.
	return q.queues[q.bufSelect.Get()&1].Pop()
}

// AppendTo pops all tasks from this queue and pushes them to another queue.
// This operation has the same semantics as repeated calls to pop.
func (q *InterruptQueue) AppendTo(other *Queue) {
	for !q.Empty() {
		q.queues[q.bufSelect.Get()&1].AppendTo(other)
	}
}
