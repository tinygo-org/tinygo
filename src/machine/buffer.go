package machine

const bufferSize = 128

//go:volatile
type volatileByte byte

// RingBuffer is ring buffer implementation inspired by post at
// https://www.embeddedrelated.com/showthread/comp.arch.embedded/77084-1.php
//
// It has some limitations currently due to how "volatile" variables that are
// members of a struct are not compiled correctly by TinyGo.
// See https://github.com/tinygo-org/tinygo/issues/151 for details.
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

// Put stores a byte in the buffer. If the buffer is already
// full, the method will return false.
func (rb *RingBuffer) Put(val byte) bool {
	if rb.Used() != bufferSize {
		rb.head++
		rb.rxbuffer[rb.head%bufferSize] = volatileByte(val)
		return true
	}
	return false
}

// Get returns a byte from the buffer. If the buffer is empty,
// the method will return a false as the second value.
func (rb *RingBuffer) Get() (byte, bool) {
	if rb.Used() != 0 {
		rb.tail++
		return byte(rb.rxbuffer[rb.tail%bufferSize]), true
	}
	return 0, false
}
