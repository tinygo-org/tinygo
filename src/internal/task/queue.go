package task

import (
	"sync/atomic"
	"unsafe"
)

const asserts = false

// Queue is a FIFO container of tasks.
// The zero value is an empty queue.
type Queue struct {
	// in is a stack used to buffer incoming tasks.
	in Stack

	// out is a singly linked list of tasks in oldest-first order.
	// Once empty, it is refilled by dumping and flipping the input stack.
	out *Task
}

// Push a task onto the queue.
// This is atomic.
func (q *Queue) Push(t *Task) {
	q.in.Push(t)
}

// Pop a task off of the queue.
// This cannot be called concurrently.
func (q *Queue) Pop() *Task {
	next := q.out
	if next == nil {
		// Dump the input stack.
		s := q.in.dump()

		// Flip it.
		var prev *Task
		for t := s.top; t != nil; {
			next := t.Next
			t.Next = prev
			prev = t
			t = next
		}
		if prev == nil {
			// The queue is empty.
			return nil
		}

		// Save it in the output list.
		next = prev
	}

	q.out = next.Next
	next.Next = nil
	return next
}

// Empty checks if the queue is empty.
// This cannot be called concurrently with Pop.
func (q *Queue) Empty() bool {
	return q.out == nil && atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&q.in.top))) == nil
}

// Stack is a LIFO container of tasks.
// The zero value is an empty stack.
// This is slightly cheaper than a queue, so it can be preferable when strict ordering is not necessary.
type Stack struct {
	top *Task
}

// Push a task onto the stack.
// This is atomic.
func (s *Stack) Push(t *Task) {
	if asserts && t.Next != nil {
		panic("runtime: pushing a task to a stack with a non-nil Next pointer")
	}
	topPtr := (*unsafe.Pointer)(unsafe.Pointer(&s.top))
doPush:
	top := atomic.LoadPointer(topPtr)
	t.Next = (*Task)(top)
	if !atomic.CompareAndSwapPointer(topPtr, top, unsafe.Pointer(t)) {
		goto doPush
	}
}

// Pop a task off of the stack.
// This is atomic.
func (s *Stack) Pop() *Task {
	topPtr := (*unsafe.Pointer)(unsafe.Pointer(&s.top))
doPop:
	top := atomic.LoadPointer(topPtr)
	if top == nil {
		return nil
	}
	t := (*Task)(top)
	next := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&t.Next)))
	if !atomic.CompareAndSwapPointer(topPtr, top, next) {
		goto doPop
	}
	t.Next = nil
	return t
}

// dump the contents of the stack to another stack.
func (s *Stack) dump() Stack {
	return Stack{
		top: (*Task)(atomic.SwapPointer(
			(*unsafe.Pointer)(unsafe.Pointer(&s.top)),
			nil,
		)),
	}
}
