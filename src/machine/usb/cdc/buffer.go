package cdc

import (
	"runtime/volatile"
)

const bufferSize = 128

// RingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type RingBuffer struct {
	rxbuffer [bufferSize]volatile.Register8
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
func (rb *RingBuffer) Put(val byte) bool {
	if rb.Used() != bufferSize {
		rb.head.Set(rb.head.Get() + 1)
		rb.rxbuffer[rb.head.Get()%bufferSize].Set(val)
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer) Get() (byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		return rb.rxbuffer[rb.tail.Get()%bufferSize].Get(), true
	}
	return 0, false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}

const bufferSize2 = 8

// RingBuffer2 is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type RingBuffer2 struct {
	rxbuffer [bufferSize2]struct {
		buf  [64]byte
		size int
	}
	head volatile.Register8
	tail volatile.Register8
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer2() *RingBuffer2 {
	return &RingBuffer2{}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer2) Used() uint8 {
	return uint8(rb.head.Get() - rb.tail.Get())
}

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer2) Put(val []byte) bool {
	if rb.Used() == bufferSize2 {
		return false
	}

	if rb.Used() == 0 {
		rb.head.Set(rb.head.Get() + 1)
		rb.rxbuffer[rb.head.Get()%bufferSize2].size = 0
	}
	buf := &rb.rxbuffer[rb.head.Get()%bufferSize2]

	for i := 0; i < len(val); i++ {
		if buf.size == 64 {
			// next
			rb.head.Set(rb.head.Get() + 1)
			buf = &rb.rxbuffer[rb.head.Get()%bufferSize2]
			rb.rxbuffer[rb.head.Get()%bufferSize2].size = 0
		}
		buf.buf[buf.size] = val[i]
		buf.size++
	}
	return true
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer2) Get() ([]byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		size := rb.rxbuffer[rb.tail.Get()%bufferSize2].size
		return rb.rxbuffer[rb.tail.Get()%bufferSize2].buf[:size], true
	}
	return nil, false
}

// Clear resets the head and tail pointer to zero.
func (rb *RingBuffer2) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}
