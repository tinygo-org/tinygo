//go:build tinygo

package escapeallocs

import (
	"encoding/binary"
	"testing"
)

var sinkU uint64

// sum reslices b in loops, so the slice is carried around as a whole.
func sum(b []byte) uint64 {
	var h uint64
	if len(b) >= 32 {
		var v1, v2 uint64
		for len(b) >= 32 {
			v1 += binary.LittleEndian.Uint64(b[0:8:len(b)])
			v2 += binary.LittleEndian.Uint64(b[8:16:len(b)])
			b = b[32:len(b):len(b)]
		}
		h = v1 ^ v2
	}
	for ; len(b) >= 8; b = b[8:] {
		h += binary.LittleEndian.Uint64(b[:8])
	}
	for ; len(b) > 0; b = b[1:] {
		h += uint64(b[0])
	}
	return h
}

func sumTwice(a, b []byte) uint64 { return sum(a) + sum(b) }

// A constant string converted to []byte and passed to a function that only reads
// it, while the function keeps the slice as a whole across loops.
func TestStringToBytesResliceInCallee(t *testing.T) {
	long := make([]byte, 100)
	sinkU = sum(long) + sumTwice(long, long[:7])
	allocs := testing.AllocsPerRun(100, func() {
		sinkU = sum([]byte("abcdefghijklmnop"))
	})
	if allocs != 0 {
		t.Errorf("got %v allocations, want 0", allocs)
	}
}
