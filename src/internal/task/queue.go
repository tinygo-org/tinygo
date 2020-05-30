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
	i := interrupt.Disable()
	if asserts && t.Next != nil {
		interrupt.Restore(i)
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
	interrupt.Restore(i)
}

// Pop a task off of the queue.
func (q *Queue) Pop() *Task {
	i := interrupt.Disable()
	t := q.head
	if t == nil {
		interrupt.Restore(i)
		return nil
	}
	q.head = t.Next
	if q.tail == t {
		q.tail = nil
	}
	t.Next = nil
	interrupt.Restore(i)
	return t
}

// Append pops the contents of another queue and pushes them onto the end of this queue.
func (q *Queue) Append(other *Queue) {
	i := interrupt.Disable()
	if q.head == nil {
		q.head = other.head
	} else {
		q.tail.Next = other.head
	}
	q.tail = other.tail
	other.head, other.tail = nil, nil
	interrupt.Restore(i)
}

// Empty checks if the queue is empty.
func (q *Queue) Empty() bool {
	i := interrupt.Disable()
	empty := q.head == nil
	interrupt.Restore(i)
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
	i := interrupt.Disable()
	if asserts && t.Next != nil {
		interrupt.Restore(i)
		panic("runtime: pushing a task to a stack with a non-nil Next pointer")
	}
	s.top, t.Next = t, s.top
	interrupt.Restore(i)
}

// Pop a task off of the stack.
func (s *Stack) Pop() *Task {
	i := interrupt.Disable()
	t := s.top
	if t != nil {
		s.top = t.Next
		t.Next = nil
	}
	interrupt.Restore(i)
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
	i := interrupt.Disable()
	head := s.top
	s.top = nil
	q := Queue{
		head: head,
		tail: head.tail(),
	}
	interrupt.Restore(i)
	return q
}
