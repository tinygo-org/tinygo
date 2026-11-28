package machine

import "internal/ringbuffer"

// RingBuffer keeps the machine.RingBuffer name for existing users.
type RingBuffer = ringbuffer.Buffer

// NewRingBuffer returns a new ring buffer.
func NewRingBuffer() *RingBuffer {
	return &RingBuffer{}
}
