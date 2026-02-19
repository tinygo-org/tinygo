package cdc

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

// --- Basic Functionality ---

func TestRing512_PutPeekDiscard(t *testing.T) {
	var r ring512
	data := []byte("hello world")
	if !r.Put(data) {
		t.Fatal("Put failed on empty buffer")
	}
	got := r.Peek()
	if !bytes.Equal(got, data) {
		t.Fatalf("Peek = %q, want %q", got, data)
	}
	if r.Used() != uint32(len(data)) {
		t.Fatalf("Used = %d, want %d", r.Used(), len(data))
	}
	r.Discard(uint32(len(data)))
	if r.Used() != 0 {
		t.Fatalf("Used after full discard = %d, want 0", r.Used())
	}
	if r.Peek() != nil {
		t.Fatalf("Peek after full discard = %v, want nil", r.Peek())
	}
}

func TestRing512_Reset(t *testing.T) {
	var r ring512
	r.Put([]byte("data"))
	r.Reset()
	if r.Used() != 0 {
		t.Fatalf("Used after Reset = %d", r.Used())
	}
	if r.Free() != 512 {
		t.Fatalf("Free after Reset = %d", r.Free())
	}
}

func TestRing512_PutEmpty(t *testing.T) {
	var r ring512
	if !r.Put(nil) {
		t.Fatal("Put nil should succeed")
	}
	if !r.Put([]byte{}) {
		t.Fatal("Put empty slice should succeed")
	}
	if r.Used() != 0 {
		t.Fatalf("Used = %d after empty puts", r.Used())
	}
}

func TestRing512_PutFull(t *testing.T) {
	var r ring512
	data := make([]byte, 512)
	for i := range data {
		data[i] = byte(i)
	}
	if !r.Put(data) {
		t.Fatal("Put 512 bytes failed on empty buffer")
	}
	if r.Free() != 0 {
		t.Fatalf("Free after filling = %d", r.Free())
	}
	if r.Put([]byte{0x42}) {
		t.Fatal("Put on full buffer should fail")
	}
	got := r.Peek()
	if !bytes.Equal(got, data) {
		t.Fatalf("Peek full buffer: got len %d, want 512", len(got))
	}
}

func TestRing512_PutExactFit(t *testing.T) {
	var r ring512
	data := make([]byte, 512)
	for i := range data {
		data[i] = byte(i)
	}
	if !r.Put(data) {
		t.Fatal("Put exact fit failed")
	}
	if r.Used() != 512 {
		t.Fatalf("Used = %d, want 512", r.Used())
	}
	r.Discard(512)
	if r.Used() != 0 {
		t.Fatal("buffer not empty after discard all")
	}
}

// --- Full buffer with off != 0 (was the critical bug in sentinel design) ---

func TestRing512_FullBufferWrapped(t *testing.T) {
	var r ring512

	// Advance tail to non-zero position.
	r.Put(make([]byte, 200))
	r.Discard(100) // tail=100, head=200, used=100

	// Fill remaining space exactly.
	free := r.Free()
	if free != 412 {
		t.Fatalf("Free = %d, want 412", free)
	}
	fill := make([]byte, free)
	for i := range fill {
		fill[i] = byte(i)
	}
	if !r.Put(fill) {
		t.Fatalf("Put(%d) into %d free space failed", free, free)
	}

	if r.Used() != 512 {
		t.Fatalf("Used = %d, want 512 (full)", r.Used())
	}
	if r.Free() != 0 {
		t.Fatalf("Free = %d, want 0 (full)", r.Free())
	}

	// Drain and verify all data recoverable.
	var drained []byte
	for r.Used() > 0 {
		p := r.Peek()
		if len(p) == 0 {
			t.Fatal("Used > 0 but Peek returned nil")
		}
		drained = append(drained, p...)
		r.Discard(uint32(len(p)))
	}
	if uint32(len(drained)) != 512 {
		t.Fatalf("drained %d bytes, want 512", len(drained))
	}
}

// --- Wrapping Tests ---

func TestRing512_Wrap(t *testing.T) {
	var r ring512
	filler := make([]byte, 500)
	if !r.Put(filler) {
		t.Fatal("fill failed")
	}
	r.Discard(490) // tail=490, head=500, used=10

	// Put 30 bytes. Will wrap: 12 at end of buf + 18 at start.
	wrapData := make([]byte, 30)
	for i := range wrapData {
		wrapData[i] = byte(i + 100)
	}
	if !r.Put(wrapData) {
		t.Fatal("wrapped Put failed")
	}
	if r.Used() != 40 {
		t.Fatalf("Used = %d, want 40", r.Used())
	}

	// Drain with two Peek/Discard rounds and verify data.
	var drained []byte
	for r.Used() > 0 {
		p := r.Peek()
		if len(p) == 0 {
			t.Fatal("Used > 0 but Peek returned nil")
		}
		drained = append(drained, p...)
		r.Discard(uint32(len(p)))
	}
	if uint32(len(drained)) != 40 {
		t.Fatalf("drained %d bytes, want 40", len(drained))
	}
}

func TestRing512_WrapDataIntegrity(t *testing.T) {
	var r ring512

	// Advance to position near end of buffer.
	r.Put(make([]byte, 500))
	r.Discard(500) // tail=500, head=500

	// Put data that wraps.
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	if !r.Put(data) {
		t.Fatal("wrapped put failed")
	}
	// Data occupies buf[500:512] + buf[0:88]

	var got []byte
	for r.Used() > 0 {
		p := r.Peek()
		got = append(got, p...)
		r.Discard(uint32(len(p)))
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("data integrity failure across wrap: got %v, want %v", got[:10], data[:10])
	}
}

// --- Edge Cases ---

func TestRing512_DiscardPartial(t *testing.T) {
	var r ring512
	r.Put([]byte("abcdefgh"))
	r.Discard(3)
	got := r.Peek()
	if !bytes.Equal(got, []byte("defgh")) {
		t.Fatalf("after partial discard, Peek = %q, want %q", got, "defgh")
	}
}

func TestRing512_DiscardZero(t *testing.T) {
	var r ring512
	r.Discard(0) // should not panic on empty
	r.Put([]byte("hi"))
	r.Discard(0) // should not panic on non-empty
	if r.Used() != 2 {
		t.Fatalf("Used = %d after zero discard", r.Used())
	}
}

func TestRing512_DiscardPanicOnOverread(t *testing.T) {
	var r ring512
	r.Put([]byte("hi"))
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic on over-discard, got none")
		}
	}()
	r.Discard(100)
}

func TestRing512_FreeUsedInvariant(t *testing.T) {
	var r ring512
	check := func(label string) {
		if r.Free()+r.Used() != 512 {
			t.Fatalf("%s: Free(%d) + Used(%d) != 512", label, r.Free(), r.Used())
		}
	}
	check("empty")
	r.Put(make([]byte, 200))
	check("after put 200")
	r.Discard(50)
	check("after discard 50")
	r.Put(make([]byte, 362))
	check("after fill to full")
	r.Discard(512)
	check("after drain")
}

func TestRing512_PutOversize(t *testing.T) {
	var r ring512
	if r.Put(make([]byte, 513)) {
		t.Fatal("Put(513) should fail on empty 512 buffer")
	}
	r.Put(make([]byte, 1))
	if r.Put(make([]byte, 512)) {
		t.Fatal("Put(512) should fail with 1 byte used")
	}
}

func TestRing512_MultiplePutPeekDiscard(t *testing.T) {
	var r ring512
	for i := 0; i < 2000; i++ {
		msg := []byte(fmt.Sprintf("msg%04d", i))
		if !r.Put(msg) {
			t.Fatalf("Put failed at iteration %d, Free=%d, Used=%d", i, r.Free(), r.Used())
		}
		// Drain completely each iteration.
		var got []byte
		for r.Used() > 0 {
			p := r.Peek()
			got = append(got, p...)
			r.Discard(uint32(len(p)))
		}
		if !bytes.Equal(got, msg) {
			t.Fatalf("iter %d: got %q, want %q", i, got, msg)
		}
	}
}

// TestRing512_HeadTailOverflow verifies correctness near uint32 max.
func TestRing512_HeadTailOverflow(t *testing.T) {
	var r ring512
	// Artificially set head/tail near overflow point.
	near := uint32(0xFFFFFFFF - 100)
	r.head.Store(near)
	r.tail.Store(near)

	if r.Used() != 0 {
		t.Fatalf("Used = %d, want 0", r.Used())
	}
	if r.Free() != 512 {
		t.Fatalf("Free = %d, want 512", r.Free())
	}

	// Write and read across the overflow boundary.
	for i := 0; i < 300; i++ {
		data := []byte{byte(i), byte(i + 1), byte(i + 2)}
		if !r.Put(data) {
			t.Fatalf("Put failed at iter %d (head=%d tail=%d)", i, r.head.Load(), r.tail.Load())
		}
		var got []byte
		for r.Used() > 0 {
			p := r.Peek()
			got = append(got, p...)
			r.Discard(uint32(len(p)))
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("iter %d: data mismatch: got %v want %v", i, got, data)
		}
	}

	// Head should have wrapped past 0.
	if r.head.Load() > 1000 && r.head.Load() < near {
		t.Logf("head didn't wrap as expected: %d", r.head.Load())
	}
}

// --- Concurrent SPSC Test ---

func TestRing512_SPSC(t *testing.T) {
	for trial := 0; trial < 20; trial++ {
		var r ring512
		const totalBytes = 1 << 18 // 256 KiB per trial
		produced := make([]byte, totalBytes)
		for i := range produced {
			produced[i] = byte(i + trial)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// Producer
		go func() {
			defer wg.Done()
			sent := 0
			for sent < totalBytes {
				chunkSize := 1 + rand.Intn(128)
				if sent+chunkSize > totalBytes {
					chunkSize = totalBytes - sent
				}
				if r.Put(produced[sent : sent+chunkSize]) {
					sent += chunkSize
				}
			}
		}()

		// Consumer
		consumed := make([]byte, 0, totalBytes)
		go func() {
			defer wg.Done()
			for len(consumed) < totalBytes {
				p := r.Peek()
				if len(p) == 0 {
					continue
				}
				consumed = append(consumed, p...)
				r.Discard(uint32(len(p)))
			}
		}()

		wg.Wait()
		if !bytes.Equal(consumed, produced) {
			for i := range consumed {
				if i >= len(produced) || consumed[i] != produced[i] {
					t.Fatalf("trial %d: mismatch at byte %d", trial, i)
				}
			}
			t.Fatalf("trial %d: length mismatch: got %d want %d", trial, len(consumed), len(produced))
		}
	}
}

// TestRing512_SPSCSmallChunks hammers single-byte puts to maximize
// wrap transitions and contention on the hot path.
func TestRing512_SPSCSmallChunks(t *testing.T) {
	var r ring512
	const totalBytes = 1 << 16

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < totalBytes; i++ {
			for !r.Put([]byte{byte(i)}) {
			}
		}
	}()

	consumed := make([]byte, 0, totalBytes)
	go func() {
		defer wg.Done()
		for len(consumed) < totalBytes {
			p := r.Peek()
			if len(p) == 0 {
				continue
			}
			consumed = append(consumed, p...)
			r.Discard(uint32(len(p)))
		}
	}()

	wg.Wait()
	for i, b := range consumed {
		if b != byte(i) {
			t.Fatalf("mismatch at %d: got %d want %d", i, b, byte(i))
		}
	}
}

// --- Fuzz Testing ---

// refRing is a trivially correct reference implementation for comparison.
type refRing struct {
	data []byte
}

func (r *refRing) Put(d []byte) bool {
	if len(r.data)+len(d) > 512 {
		return false
	}
	r.data = append(r.data, d...)
	return true
}

func (r *refRing) Peek() []byte { return r.data }

func (r *refRing) Discard(n uint32) {
	if uint32(len(r.data)) < n {
		panic("ref: discard overflow")
	}
	r.data = r.data[n:]
}

func (r *refRing) Used() uint32 { return uint32(len(r.data)) }
func (r *refRing) Free() uint32 { return 512 - uint32(len(r.data)) }
func (r *refRing) Reset()       { r.data = r.data[:0] }

// FuzzRing512 runs random sequences of operations against both Ring512 and
// a reference implementation, comparing results.
func FuzzRing512(f *testing.F) {
	f.Add([]byte{0, 10, 1, 5, 2, 0, 10, 1, 10})
	f.Add([]byte{0, 0})
	f.Add([]byte{0, 255, 0, 255, 1, 255, 1, 255})
	f.Add(bytes.Repeat([]byte{0, 64, 1, 64}, 50))
	// Seed that triggers full-buffer-with-wrap (killed old sentinel design).
	f.Add([]byte{0, 200, 1, 100, 0, 156, 0, 156})

	f.Fuzz(func(t *testing.T, ops []byte) {
		var ring ring512
		var ref refRing

		i := 0
		for i < len(ops) {
			op := ops[i] % 4
			i++
			if i >= len(ops) {
				break
			}
			arg := ops[i]
			i++

			switch op {
			case 0: // Put
				size := int(arg)
				if size > 512 {
					size = 512
				}
				data := make([]byte, size)
				for j := range data {
					data[j] = byte(j)
				}
				gotOK := ring.Put(data)
				refOK := ref.Put(data)
				if gotOK != refOK {
					t.Fatalf("Put(%d): ring=%v ref=%v (ringUsed=%d refUsed=%d)",
						size, gotOK, refOK, ring.Used(), ref.Used())
				}

			case 1: // Discard
				used := ring.Used()
				refUsed := ref.Used()
				if used != refUsed {
					t.Fatalf("Used mismatch before discard: ring=%d ref=%d", used, refUsed)
				}
				if used == 0 {
					continue
				}
				n := uint32(arg) % (used + 1)
				ring.Discard(n)
				ref.Discard(n)

			case 2: // Peek + verify
				ringPeek := ring.Peek()
				refPeek := ref.Peek()
				rUsed := ring.Used()
				refUsed := ref.Used()
				if rUsed != refUsed {
					t.Fatalf("Used mismatch: ring=%d ref=%d", rUsed, refUsed)
				}
				if rUsed == 0 {
					if ringPeek != nil {
						t.Fatalf("ring Peek non-nil on empty: %v", ringPeek)
					}
					continue
				}
				// Ring512 Peek returns first contiguous segment; must be a prefix.
				if uint32(len(ringPeek)) > rUsed {
					t.Fatalf("ring Peek len %d > Used %d", len(ringPeek), rUsed)
				}
				if !bytes.Equal(ringPeek, refPeek[:len(ringPeek)]) {
					t.Fatalf("Peek data mismatch")
				}

			case 3: // Invariant check
				if ring.Free()+ring.Used() != 512 {
					t.Fatalf("invariant: Free(%d)+Used(%d) != 512", ring.Free(), ring.Used())
				}
			}
		}

		// Final checks.
		if ring.Free()+ring.Used() != 512 {
			t.Fatalf("final: Free(%d)+Used(%d) != 512", ring.Free(), ring.Used())
		}
		if ring.Used() != ref.Used() {
			t.Fatalf("final Used mismatch: ring=%d ref=%d", ring.Used(), ref.Used())
		}
	})
}

// FuzzRing512_PutDrain fuzzes fill-then-drain cycles to exercise all
// wrap positions and buffer-full states.
func FuzzRing512_Op(f *testing.F) {
	const maxsz = 512
	f.Add(int16(7), int16(200), int16(50), int16(180))
	f.Add(int16(maxsz), int16(-maxsz), int16(maxsz), int16(-maxsz))

	f.Fuzz(func(t *testing.T, a, b, c, d int16) {
		rng := rand.New(rand.NewSource(int64(a + b + c + d)))
		var ring ring512
		var buf [maxsz]byte
		sizes := [...]int{int(a), int(b), int(c), int(d)}
		var testwritten, testread [maxsz * len(sizes)]byte
		nwritten := 0
		nread := 0
		currentUsed := 0
		initfree := ring.Free()
		for round, sz := range sizes {
			write := sz > 0
			if sz < 0 {
				sz = -sz
			}
			if sz > maxsz {
				sz = maxsz
			}
			free := int(ring.Free())
			used := int(ring.Used())
			if free+used != int(initfree) {
				t.Fatalf("free+used != initfree: %d+%d!=%d", free, used, initfree)
			} else if used != currentUsed {
				t.Fatalf("calculated used not match actual used returned %d!=%d", used, currentUsed)
			}
			rng.Read(buf[:sz])
			if write {
				sz = min(free, sz) // Limit write to be size of free.
				nwritten += copy(testwritten[nwritten:], buf[:sz])
				ok := ring.Put(buf[:sz])
				if !ok {
					t.Fatal("tried to put data and could not", sz)
				}
				currentUsed += sz
			} else {
				// read branch.
				sz = min(currentUsed, sz) // Limit size of operation to what is possible.
				data1 := ring.Peek()
				data1 = data1[:min(sz, len(data1))]
				nread += copy(testread[nread:], data1)
				ring.Discard(uint32(len(data1)))
				if len(data1) < sz {
					data2 := ring.Peek()
					if len(data2) <= 0 {
						t.Fatal("expected more data after first discard")
					} else if len(data2)+len(data1) < sz {
						t.Fatalf("got promised more data %d+%d<%d", len(data1), len(data2), sz)
					} else if int(ring.Used()) != currentUsed-len(data1) {
						t.Fatalf("expected new used to be old used minus read %d != %d-%d", ring.Used(), currentUsed, len(data1))
					}
					data2 = data2[:sz-len(data1)]
					nread += copy(testread[nread:], data1)
					ring.Discard(uint32(len(data2)))
				}
				currentUsed -= sz
			}
			if int(ring.Used()) != currentUsed {
				t.Fatalf("unexpected new used after read/write %d!=%d", ring.Used(), currentUsed)
			}
			if !write {
				// check data read/written match.
				testlim := min(nread, nwritten)
				if !bytes.Equal(testread[:testlim], testwritten[:testlim]) {
					t.Fatalf("round %d mismatch of data written/read", round)
				}
			}
		}
	})
}
