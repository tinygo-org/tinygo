package bytealg_test

import (
	"internal/bytealg"
	"testing"
)

func TestCountByte(t *testing.T) {
	data := make([]byte, 80)
	state := uint32(1)
	for i := range data {
		state = state*1664525 + 1013904223
		data[i] = byte(state >> 24)
	}

	for offset := 0; offset < 8; offset++ {
		for length := 0; length <= 64; length++ {
			value := data[offset : offset+length]
			text := string(value)
			for needle := 0; needle < 256; needle++ {
				want := countByteScalar(value, byte(needle))
				if got := bytealg.Count(value, byte(needle)); got != want {
					t.Fatalf("Count(offset=%d, length=%d, byte=%d) = %d, want %d",
						offset, length, needle, got, want)
				}
				if got := bytealg.CountString(text, byte(needle)); got != want {
					t.Fatalf("CountString(offset=%d, length=%d, byte=%d) = %d, want %d",
						offset, length, needle, got, want)
				}
			}
		}
	}
}

func countByteScalar(value []byte, needle byte) int {
	count := 0
	for _, b := range value {
		if b == needle {
			count++
		}
	}
	return count
}
