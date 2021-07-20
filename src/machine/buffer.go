package machine

import (
	"runtime/volatile"
)

const defaultRingBufferSize = 128

// RingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type RingBuffer struct {
	rxbuffer []volatile.Register8
	head     volatile.Register16
	tail     volatile.Register16
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer(size ...int) *RingBuffer {
	sz := defaultRingBufferSize
	if len(size) > 0 {
		sz = size[0]
	}
	return &RingBuffer{
		rxbuffer: make([]volatile.Register8, sz),
	}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer) Used() uint16 {
	return uint16(rb.head.Get() - rb.tail.Get())
}

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer) Put(val byte) bool {
	if rb.Used() != uint16(len(rb.rxbuffer)) {
		rb.head.Set(rb.head.Get() + 1)
		rb.rxbuffer[rb.head.Get()%uint16(len(rb.rxbuffer))].Set(val)
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer) Get() (byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		return rb.rxbuffer[rb.tail.Get()%uint16(len(rb.rxbuffer))].Get(), true
	}
	return 0, false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}
