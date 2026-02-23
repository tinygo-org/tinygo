package cdc

import "sync/atomic"

// ring512 is an interrupt/concurrent-safe ring buffer for a single-producer,
// single-consumer (SPSC) pair. The writer calls Put, the reader calls
// Peek/Discard. Reset may only be called when neither side is active.
//
// Implementation uses monotonic counters (head/tail) instead of bounded
// offsets. Unsigned subtraction (head - tail) always yields
// correct used count regardless of uint32 overflow.
type ring512 struct {
	buf [ringBufLen]byte // power of 2 so compiler can optimize & mask.
	// head counts total bytes written. Only the writer stores to head.
	head atomic.Uint32
	// tail counts total bytes read. Only the reader stores to tail.
	tail atomic.Uint32
}

const (
	ringBufLen = 512
	ringMask   = ringBufLen - 1 // 0x1FF
)

// Reset empties the ring buffer. Must not be called concurrently with
// Put, Peek, or Discard.
func (r *ring512) Reset() {
	r.head.Store(0)
	r.tail.Store(0)
}

// Free returns number of bytes that can be written via Put.
func (r *ring512) Free() uint32 {
	return ringBufLen - r.Used()
}

// Used returns number of bytes ready to be peeked/discarded.
func (r *ring512) Used() uint32 {
	return r.head.Load() - r.tail.Load()
}

// Peek returns contiguous views into the readable portions of the buffer
// without advancing the read position. When data wraps around the end of
// the internal buffer, two segments are returned. Second data2 is nil on fully contiguous buffer.
// Returns nil,nil when empty.
func (r *ring512) Peek() (data1, data2 []byte) {
	head, tail := r.lims()
	used := head - tail
	if used == 0 {
		return nil, nil
	}
	pos := tail & ringMask
	contig := ringBufLen - pos
	if contig >= used {
		return r.buf[pos : pos+used], nil
	}
	return r.buf[pos:], r.buf[:used-contig]
}

// Discard marks numBytes as read, advancing the read position.
// Panics if numBytes exceeds Used (indicates a race violating SPSC)
func (r *ring512) Discard(numBytes uint32) {
	if numBytes == 0 {
		return
	}
	head, tail := r.lims()
	used := head - tail
	if numBytes > used {
		panic("ring: discard exceeds used")
	}
	r.tail.Store(tail + numBytes)
}

// Put writes data into the ring buffer. Returns true if all data was
// written, false if insufficient free space (no partial writes).
func (r *ring512) Put(data []byte) bool {
	wlen := uint32(len(data))
	if wlen == 0 {
		return true
	}
	head, tail := r.lims()
	used := head - tail
	free := uint32(ringBufLen) - used
	if wlen > free {
		return false
	}
	pos := head & ringMask
	n := uint32(copy(r.buf[pos:], data))
	if n < wlen {
		copy(r.buf[:], data[n:])
	}
	r.head.Store(head + wlen)
	return true
}

func (r *ring512) lims() (head, tail uint32) {
	return r.head.Load(), r.tail.Load()
}
