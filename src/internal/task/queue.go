package task

import "runtime/interrupt"

const asserts = false

// Queue is a FIFO container of tasks.
// The zero value is an empty queue.
type Queue struct {
	head, tail *Task
}

// Push a task onto the queue.
func (q *Queue) Push(t *Task) {
	mask := lockAtomics()
	if asserts && t.Next != nil {
		unlockAtomics(mask)
		panic("runtime: pushing a task to a queue with a non-nil Next pointer")
	}
	if q.tail != nil {
		q.tail.Next = t
	}
	q.tail = t
	t.Next = nil
	if q.head == nil {
		q.head = t
	}
	unlockAtomics(mask)
}

// Pop a task off of the queue.
func (q *Queue) Pop() *Task {
	mask := lockAtomics()
	t := q.head
	if t == nil {
		unlockAtomics(mask)
		return nil
	}
	q.head = t.Next
	if q.tail == t {
		q.tail = nil
	}
	t.Next = nil
	unlockAtomics(mask)
	return t
}

// Append pops the contents of another queue and pushes them onto the end of this queue.
func (q *Queue) Append(other *Queue) {
	mask := lockAtomics()
	if q.head == nil {
		q.head = other.head
	} else {
		q.tail.Next = other.head
	}
	q.tail = other.tail
	other.head, other.tail = nil, nil
	unlockAtomics(mask)
}

// Empty checks if the queue is empty.
func (q *Queue) Empty() bool {
	mask := lockAtomics()
	empty := q.head == nil
	unlockAtomics(mask)
	return empty
}

// Stack is a LIFO container of tasks.
// The zero value is an empty stack.
// This is slightly cheaper than a queue, so it can be preferable when strict ordering is not necessary.
type Stack struct {
	top *Task
}

// Push a task onto the stack.
func (s *Stack) Push(t *Task) {
	mask := lockAtomics()
	if asserts && t.Next != nil {
		unlockAtomics(mask)
		panic("runtime: pushing a task to a stack with a non-nil Next pointer")
	}
	s.top, t.Next = t, s.top
	unlockAtomics(mask)
}

// Pop a task off of the stack.
func (s *Stack) Pop() *Task {
	mask := lockAtomics()
	t := s.top
	if t != nil {
		s.top = t.Next
		t.Next = nil
	}
	unlockAtomics(mask)
	return t
}

// tail follows the chain of tasks.
// If t is nil, returns nil.
// Otherwise, returns the task in the chain where the Next field is nil.
func (t *Task) tail() *Task {
	if t == nil {
		return nil
	}
	for t.Next != nil {
		t = t.Next
	}
	return t
}

// Queue moves the contents of the stack into a queue.
// Elements can be popped from the queue in the same order that they would be popped from the stack.
func (s *Stack) Queue() Queue {
	mask := lockAtomics()
	head := s.top
	s.top = nil
	q := Queue{
		head: head,
		tail: head.tail(),
	}
	unlockAtomics(mask)
	return q
}

// Use runtime.lockAtomics and runtime.unlockAtomics so that Queue and Stack
// work correctly even on multicore systems. These functions are normally used
// to implement atomic operations, but the same spinlock can also be used for
// Queue/Stack operations which are very fast.
// These functions are just plain old interrupt disable/restore on non-multicore
// systems.

//go:linkname lockAtomics runtime.lockAtomics
func lockAtomics() interrupt.State

//go:linkname unlockAtomics runtime.unlockAtomics
func unlockAtomics(mask interrupt.State)
