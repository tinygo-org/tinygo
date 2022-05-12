package usb

import (
	"errors"
	"runtime/volatile"
)

const QueueSize = 8

type Queue struct {
	fifo [QueueSize]uint8
	tail volatile.Register32
	head volatile.Register32
}

var (
	ErrReadBuffer  = errors.New("cannot copy into read buffer")
	ErrWriteBuffer = errors.New("cannot copy from write buffer")
	ErrQueueEmpty  = errors.New("data queue empty") // Read Error
	ErrQueueFull   = errors.New("data queue full")  // Write error
)

// Reset discards all buffered data.
//go:inline
func (q *Queue) Reset() {
	for i := range q.fifo {
		q.fifo[i] = 0
	}
	q.tail.Set(0)
	q.head.Set(0)
}

//go:inline
func (q *Queue) Len() int {
	return int(q.tail.Get() - q.head.Get())
}

//go:inline
func (q *Queue) Cap() int {
	return cap(q.fifo)
}

func (q *Queue) Read(data []uint8) (int, error) {

	less := uint32(len(data))
	if less == 0 {
		return 0, ErrReadBuffer
	} // nothing to copy into

	head := q.head.Get()
	used := q.tail.Get() - head

	if used == 0 {
		return 0, ErrQueueEmpty
	} // empty queue

	if less > used {
		less = used
	} // only get from used space

	for i := uint32(0); i < less; i++ {
		data[i] = q.fifo[head%QueueSize]
		head++
	}
	q.head.Set(head)

	return int(less), nil
}

func (q *Queue) Write(data []uint8) (int, error) {

	more := uint32(len(data))
	if more == 0 {
		return 0, ErrWriteBuffer
	} // nothing to copy from

	tail := q.tail.Get()
	used := tail - q.head.Get()

	if used == QueueSize {
		return 0, ErrQueueFull
	} // full queue

	if used+more > QueueSize {
		more = QueueSize - used
	} // only put to unused space

	for i := uint32(0); i < more; i++ {
		q.fifo[tail%QueueSize] = data[i]
		tail++
	}
	q.tail.Set(tail)

	return int(more), nil
}

func (q *Queue) Deq() (uint8, bool) {

	tail := q.head.Get()
	if tail == q.tail.Get() {
		return 0, false
	} // empty queue

	data := q.fifo[tail%QueueSize]
	q.head.Set(tail + 1)

	return data, true
}

func (q *Queue) Front() (uint8, bool) {

	tail := q.head.Get()
	if tail == q.tail.Get() {
		return 0, false
	} // empty queue

	return q.fifo[tail%QueueSize], true
}

func (q *Queue) Enq(data uint8) bool {

	head := q.tail.Get()
	if head-q.head.Get() == QueueSize {
		return false
	} // full queue

	q.fifo[head%QueueSize] = data
	q.tail.Set(head + 1)

	return true
}
