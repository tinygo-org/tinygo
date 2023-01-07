package joystick

import (
	"runtime/volatile"
)

const bufferSize = 32

// RingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type RingBuffer struct {
	rxbuffer [bufferSize][16]byte
	head     volatile.Register8
	tail     volatile.Register8
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer() *RingBuffer {
	return &RingBuffer{}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer) Used() uint8 {
	return uint8(rb.head.Get() - rb.tail.Get())
}

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer) Put(val []byte) bool {
	if rb.Used() != bufferSize {
		rb.head.Set(rb.head.Get() + 1)
		copy(rb.rxbuffer[rb.head.Get()%bufferSize][:], val)
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer) Get() ([]byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		return rb.rxbuffer[rb.tail.Get()%bufferSize][:], true
	}
	return nil, false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}
