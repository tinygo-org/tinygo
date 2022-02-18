package usb

import (
	"errors"
	"runtime/volatile"
)

type QueueFullDiscardMode uint8

const (
	QueueFullDiscardLast  QueueFullDiscardMode = iota // Drop incoming data
	QueueFullDiscardFirst                             // Drop outgoing data
)

type Queue struct {
	mode QueueFullDiscardMode
	size volatile.Register32
	fifo *[]uint8
	tail volatile.Register32 // New elements are enqueued at index tail
	head volatile.Register32 // Oldest element in queue is at index head
}

var (
	ErrReadBuffer  = errors.New("cannot copy into read buffer")
	ErrWriteBuffer = errors.New("cannot copy from write buffer")
	ErrQueueEmpty  = errors.New("data queue empty") // Read Error
	ErrQueueFull   = errors.New("data queue full")  // Write error
	ErrQueueNoMode = errors.New("unknown FIFO copy mode")
)

// Init initializes the receiver queue's backing data store with the given byte
// slice fifo and logical capacity size. If size is greater than the slice's
// physical length, uses the slice's physical length.
func (q *Queue) Init(fifo *[]uint8, size int, mode QueueFullDiscardMode) {
	q.mode = mode
	q.fifo = fifo
	q.Reset(size)
}

// Reset discards all buffered data and sets the FIFO logical capacity.
// If size is less than 0 or greater than FIFO physical length, uses FIFO
// physical length.
//go:inline
func (q *Queue) Reset(size int) {
	if phy := len(*q.fifo); size < 0 || size > phy {
		size = phy
	}
	q.size.Set(uint32(size))
	q.tail.Set(0)
	q.head.Set(0)
}

// Cap returns the logical capacity of the receiver FIFO.
//go:inline
func (q *Queue) Cap() int {
	return int(q.size.Get())
}

// Len returns the number of elements enqueued in the receiver FIFO.
//go:inline
func (q *Queue) Len() int {
	return int(q.tail.Get() - q.head.Get())
}

// Rem returns the number of elements not enqueued in the receiver FIFO.
//go:inline
func (q *Queue) Rem() int {
	return q.Cap() - q.Len()
}

// Deq dequeues and returns the element at the front of the receiver FIFO and true.
// If the FIFO is empty and no element was dequeued, returns 0 and false.
func (q *Queue) Deq() (uint8, bool) {

	head := q.head.Get()
	if head == q.tail.Get() {
		return 0, false
	} // empty queue

	data := (*q.fifo)[head%q.size.Get()]
	q.head.Set(head + 1)

	return data, true
}

// Enq enqueues the given element data at the back of the receiver FIFO and
// returns true.
// If the FIFO is full and no element can be enqueued, returns false.
//
// TODO(ardnew): Document both operations based on receiver's QueueFullMode.
func (q *Queue) Enq(data uint8) bool {

	tail := q.tail.Get()
	head := q.head.Get()
	if tail-head == q.size.Get() {
		switch q.mode {
		case QueueFullDiscardLast:
			// drop incoming data
			return false
		case QueueFullDiscardFirst:
			// drop outgoing data
			q.head.Set(head + 1)
		}
	} // full queue

	(*q.fifo)[tail%q.size.Get()] = data
	q.tail.Set(tail + 1)

	return true
}

// Read implements the io.Reader interface. It dequeues min(q.Len(), len(data))
// elements from the receiver FIFO into the given slice data.
// If len(data) equals 0, returns 0 and ErrReadBuffer.
// Otherwise, if q.Len() equals 0, returns 0 and ErrQueueEmpty.
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
		data[i] = (*q.fifo)[head%q.size.Get()]
		head++
	}
	q.head.Set(head)

	return int(less), nil
}

// Write implements the io.Writer interface. It enqueues min(q.Rem(), len(data))
// elements from the given slice data into the receiver FIFO.
// If len(data) equals 0, returns 0 and ErrWriteBuffer.
// Otherwise, if q.Rem() equals 0, returns 0 and ErrQueueFull.
//
// TODO(ardnew): Document both operations based on receiver's QueueFullMode.
func (q *Queue) Write(data []uint8) (int, error) {

	more := uint32(len(data))

	// Nothing to copy from is an error regardless of mode.
	if more == 0 {
		return 0, ErrWriteBuffer
	}

	switch q.mode {
	case QueueFullDiscardLast:
		// drop incoming data

		tail := q.tail.Get()
		used := tail - q.head.Get()

		// Full queue, cannot add any data.
		if used == q.size.Get() {
			return 0, ErrQueueFull
		}

		// xOnly put to unused space.
		if used+more > q.size.Get() {
			more = q.size.Get() - used
		}

		// Copy a potentially-limited number of elements from data, depending on the
		// current length of FIFO.
		for i := uint32(0); i < more; i++ {
			(*q.fifo)[tail%q.size.Get()] = data[i]
			tail++
		}
		q.tail.Set(tail)

		return int(more), nil

	case QueueFullDiscardFirst:
		// drop outgoing data

		// Trying to write more data than the FIFO will hold will simply overwrite
		// some of the given data, so there is no point writing that data.
		from := uint32(0)
		if more >= q.size.Get() {
			// Begin copying only the data that will be kept.
			from = more - q.size.Get()
			// We can fill the entire FIFO.
			more = q.size.Get()
			// Reset the indices
			q.head.Set(0)
			q.tail.Set(0)
		}

		tail := q.tail.Get()
		used := tail - q.head.Get()

		// Make space for incoming data by discarding only as many FIFO elements as
		// is necessary to store incoming data.
		if used+more > q.size.Get() {
			q.head.Set(tail + more - q.size.Get())
		}

		// Copy a potentially-limited number of elements from data, depending on the
		// current length of FIFO.
		for i := uint32(0); i < more; i++ {
			(*q.fifo)[tail%q.size.Get()] = data[i]
			tail++
		}

	}
}

// Front returns the next element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns 0 and false.
func (q *Queue) Front() (uint8, bool) {

	head := q.head.Get()
	if head == q.tail.Get() {
		return 0, false
	} // empty queue

	return (*q.fifo)[head%q.size.Get()], true
}

// Back returns the last element that would be dequeued from the receiver FIFO
// and true.
// If the FIFO is empty and no element would be dequeued, returns 0 and false.
func (q *Queue) Back() (uint8, bool) {

	tail := q.tail.Get()
	if tail == q.head.Get() {
		return 0, false
	} // empty queue

	return (*q.fifo)[(tail-1)%q.size.Get()], true
}

// index returns an index into the receiver FIFO based on sign and magnitude of i:
// 1. If i is greater than or equal to zero and less then q.Len(), returns the
//    (i+1)'th element that would be dequeued from the receiver FIFO and true.
// 2. Otherwise, if i is negative and -i is less than or equal to q.Len(), returns
//    the -(i+1)'th from the last element that would be dequeued from the receiver
//    FIFO and true.
// 3. Otherwise, returns 0 and false.
func (q *Queue) index(i int) (int, bool) {
	if n := q.Len(); i < 0 {
		if -i <= n {
			return (int(q.tail.Get()) + i) % int(q.size.Get()), true
		}
	} else {
		if i < n {
			return (int(q.head.Get()) + i) % int(q.size.Get()), true
		}
	}
	return 0, false
}

// Get returns the value of an element in the receiver FIFO, offset by i from the
// front of the queue if i is positive, or from the back of the queue if i is
// negative. For example:
//  	Get(0)  == Get(-Len())  == Front(), and
// 		Get(-1) == Get(Len()-1) == Back().
// If the offset is beyond queue boundaries, returns 0 and false.
func (q *Queue) Get(i int) (uint8, bool) {

	if n, ok := q.index(i); ok {
		return (*q.fifo)[n], true
	}
	return 0, false
}

// Set modifies the value of an element in the receiver FIFO.
// Set uses the same logic as Get to select an element in the FIFO.
func (q *Queue) Set(i int, data uint8) bool {

	if n, ok := q.index(i); ok {
		(*q.fifo)[n] = data
		return true
	}
	return false
}
