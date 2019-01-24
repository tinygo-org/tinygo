package machine

const bufferSize = 128

//go:volatile
type volatileByte byte

// RingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php

type RingBuffer struct {
	rxbuffer [bufferSize]volatileByte
	head     volatileByte
	tail     volatileByte
}

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer() *RingBuffer {
	return &RingBuffer{}
}

// Used returns how many bytes in buffer have been used.
func (rb *RingBuffer) Used() uint8 {
	return uint8(rb.head - rb.tail)
}

// Put stores a byte in the buffer.
func (rb *RingBuffer) Put(val byte) {
	if rb.Used() != bufferSize {
		rb.head++
		rb.rxbuffer[rb.head%bufferSize] = volatileByte(val)
	}
}

// Get returns a byte from the buffer.
func (rb *RingBuffer) Get() byte {
	if rb.Used() != 0 {
		rb.tail++
		return byte(rb.rxbuffer[rb.tail%bufferSize])
	}
	return 0
}
