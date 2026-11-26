package ringbuffer_test

import (
	"internal/ringbuffer"
	"testing"
)

// fill puts bytes 0, 1, 2 and so on until the buffer is full and returns the count.
func fill(t *testing.T, rb *ringbuffer.Buffer) int {
	n := 0
	for rb.Put(byte(n)) {
		n++
		if n > 256 {
			t.Fatal("Put never reported a full buffer")
		}
	}
	return n
}

func TestFullEmpty(t *testing.T) {
	var rb ringbuffer.Buffer
	if _, ok := rb.Get(); ok {
		t.Fatal("Get on empty buffer returned ok")
	}
	size := fill(t, &rb)
	if size == 0 {
		t.Fatal("Put on empty buffer failed")
	}
	if int(rb.Used()) != size {
		t.Fatalf("Used() = %d, want %d", rb.Used(), size)
	}
	for i := 0; i < size; i++ {
		b, ok := rb.Get()
		if !ok || b != byte(i) {
			t.Fatalf("Get = %d, %v, want %d, true", b, ok, i)
		}
	}
	if _, ok := rb.Get(); ok {
		t.Fatal("Get on drained buffer returned ok")
	}
}

func TestClear(t *testing.T) {
	var rb ringbuffer.Buffer
	rb.Put(1)
	rb.Put(2)
	rb.Clear()
	if rb.Used() != 0 {
		t.Fatalf("Used() = %d after Clear, want 0", rb.Used())
	}
	if _, ok := rb.Get(); ok {
		t.Fatal("Get after Clear returned ok")
	}
}

// TestWrap moves head and tail past 255 so the uint8 indexes wrap.
func TestWrap(t *testing.T) {
	var rb ringbuffer.Buffer
	var in, out byte
	for round := 0; round < 1000; round++ {
		for n := round % 7; n > 0 && rb.Put(in); n-- {
			in++
		}
		for n := round % 5; n > 0; n-- {
			b, ok := rb.Get()
			if !ok {
				break
			}
			if b != out {
				t.Fatalf("round %d: Get = %d, want %d", round, b, out)
			}
			out++
		}
	}
}
