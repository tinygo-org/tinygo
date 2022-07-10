package cdc

import (
	"runtime/volatile"
)

const rxRingBufferSize = 128

// rxRingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type rxRingBuffer struct {
	buffer [rxRingBufferSize]volatile.Register8
	head   volatile.Register8
	tail   volatile.Register8
}

// NewRxRingBuffer returns a new ring buffer.
func NewRxRingBuffer() *rxRingBuffer {
	return &rxRingBuffer{}
}

// Used returns how many bytes in buffer have been used.
func (rb *rxRingBuffer) Used() uint8 {
	return uint8(rb.head.Get() - rb.tail.Get())
}

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *rxRingBuffer) Put(val byte) bool {
	if rb.Used() != rxRingBufferSize {
		rb.head.Set(rb.head.Get() + 1)
		rb.buffer[rb.head.Get()%rxRingBufferSize].Set(val)
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *rxRingBuffer) Get() (byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		return rb.buffer[rb.tail.Get()%rxRingBufferSize].Get(), true
	}
	return 0, false
}

// Clear resets the head and tail pointer to zero.
func (rb *rxRingBuffer) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}

const txRingBufferSize = 8

// txRingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
type txRingBuffer struct {
	buffer [txRingBufferSize]struct {
		buf  [64]byte
		size int
	}
	head volatile.Register8
	tail volatile.Register8
}

// NewTxRingBuffer returns a new ring buffer.
func NewTxRingBuffer() *txRingBuffer {
	return &txRingBuffer{}
}

// Used returns how many bytes in buffer have been used.
func (rb *txRingBuffer) Used() uint8 {
	return uint8(rb.head.Get() - rb.tail.Get())
}

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *txRingBuffer) Put(val []byte) bool {
	if rb.Used() == txRingBufferSize {
		return false
	}

	if rb.Used() == 0 {
		rb.head.Set(rb.head.Get() + 1)
		rb.buffer[rb.head.Get()%txRingBufferSize].size = 0
	}
	buf := &rb.buffer[rb.head.Get()%txRingBufferSize]

	for i := 0; i < len(val); i++ {
		if buf.size == 64 {
			// next
			// TODO: Make sure that data is not corrupted even when the buffer is full
			rb.head.Set(rb.head.Get() + 1)
			buf = &rb.buffer[rb.head.Get()%txRingBufferSize]
			rb.buffer[rb.head.Get()%txRingBufferSize].size = 0
		}
		buf.buf[buf.size] = val[i]
		buf.size++
	}
	return true
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *txRingBuffer) Get() ([]byte, bool) {
	if rb.Used() != 0 {
		rb.tail.Set(rb.tail.Get() + 1)
		size := rb.buffer[rb.tail.Get()%txRingBufferSize].size
		return rb.buffer[rb.tail.Get()%txRingBufferSize].buf[:size], true
	}
	return nil, false
}

// Clear resets the head and tail pointer to zero.
func (rb *txRingBuffer) Clear() {
	rb.head.Set(0)
	rb.tail.Set(0)
}
