package cdc

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

// peekAll returns all readable data by concatenating both Peek segments.
func peekAll(r *ring512) []byte {
	d1, d2 := r.Peek()
	if len(d2) == 0 {
		return d1
	}
	out := make([]byte, len(d1)+len(d2))
	copy(out, d1)
	copy(out[len(d1):], d2)
	return out
}

// drain reads all data from the ring, verifying Peek length matches Used.
func drain(t *testing.T, r *ring512) []byte {
	t.Helper()
	d1, d2 := r.Peek()
	n := uint32(len(d1) + len(d2))
	if n != r.Used() {
		t.Fatalf("Peek returned %d bytes but Used()=%d", n, r.Used())
	}
	var out []byte
	out = append(out, d1...)
	out = append(out, d2...)
	r.Discard(n)
	return out
}

// --- Basic Functionality ---

func TestRing512_PutPeekDiscard(t *testing.T) {
	var r ring512
	data := []byte("hello world")
	if !r.Put(data) {
		t.Fatal("Put failed on empty buffer")
	}
	got := peekAll(&r)
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
	d1, d2 := r.Peek()
	if d1 != nil || d2 != nil {
		t.Fatalf("Peek after full discard = (%v, %v), want (nil, nil)", d1, d2)
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
	got := peekAll(&r)
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

// --- Full buffer with wrapped position ---

func TestRing512_FullBufferWrapped(t *testing.T) {
	var r ring512

	r.Put(make([]byte, 200))
	r.Discard(100) // tail=100, head=200, used=100

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
	drained := drain(t, &r)
	if len(drained) != 512 {
		t.Fatalf("drained %d bytes, want 512", len(drained))
	}
}

// --- Wrapping Tests ---

func TestRing512_Wrap(t *testing.T) {
	var r ring512
	r.Put(make([]byte, 500))
	r.Discard(490) // tail=490, head=500, used=10

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

	d1, d2 := r.Peek()
	if len(d1)+len(d2) != 40 {
		t.Fatalf("Peek total = %d, want 40", len(d1)+len(d2))
	}
	if d2 == nil {
		t.Fatal("expected wrapped data in d2")
	}
	drained := drain(t, &r)
	if len(drained) != 40 {
		t.Fatalf("drained %d bytes, want 40", len(drained))
	}
}

func TestRing512_WrapDataIntegrity(t *testing.T) {
	var r ring512
	r.Put(make([]byte, 500))
	r.Discard(500)

	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	if !r.Put(data) {
		t.Fatal("wrapped put failed")
	}
	got := drain(t, &r)
	if !bytes.Equal(got, data) {
		t.Fatal("data integrity failure across wrap")
	}
}

// --- Edge Cases ---

func TestRing512_DiscardPartial(t *testing.T) {
	var r ring512
	r.Put([]byte("abcdefgh"))
	r.Discard(3)
	got := peekAll(&r)
	if !bytes.Equal(got, []byte("defgh")) {
		t.Fatalf("after partial discard, Peek = %q, want %q", got, "defgh")
	}
}

func TestRing512_DiscardZero(t *testing.T) {
	var r ring512
	r.Discard(0)
	r.Put([]byte("hi"))
	r.Discard(0)
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
		got := drain(t, &r)
		if !bytes.Equal(got, msg) {
			t.Fatalf("iter %d: got %q, want %q", i, got, msg)
		}
	}
}

func TestRing512_HeadTailOverflow(t *testing.T) {
	var r ring512
	near := uint32(0xFFFFFFFF - 100)
	r.head.Store(near)
	r.tail.Store(near)

	if r.Used() != 0 || r.Free() != 512 {
		t.Fatalf("Used=%d Free=%d, want 0/512", r.Used(), r.Free())
	}

	for i := 0; i < 300; i++ {
		data := []byte{byte(i), byte(i + 1), byte(i + 2)}
		if !r.Put(data) {
			t.Fatalf("Put failed at iter %d (head=%d tail=%d)", i, r.head.Load(), r.tail.Load())
		}
		got := drain(t, &r)
		if !bytes.Equal(got, data) {
			t.Fatalf("iter %d: data mismatch", i)
		}
	}
}

// --- Peek two-segment tests ---

func TestRing512_PeekNoWrap(t *testing.T) {
	var r ring512
	r.Put([]byte("hello"))
	d1, d2 := r.Peek()
	if !bytes.Equal(d1, []byte("hello")) {
		t.Fatalf("d1 = %q, want %q", d1, "hello")
	}
	if d2 != nil {
		t.Fatalf("d2 = %v, want nil", d2)
	}
}

func TestRing512_PeekWrapped(t *testing.T) {
	var r ring512
	r.Put(make([]byte, 508))
	r.Discard(508) // tail=508, head=508

	data := []byte("abcdefghij") // 10 bytes: 4 at end, 6 at start
	if !r.Put(data) {
		t.Fatal("put failed")
	}
	d1, d2 := r.Peek()
	if len(d1) != 4 {
		t.Fatalf("d1 len = %d, want 4", len(d1))
	}
	if len(d2) != 6 {
		t.Fatalf("d2 len = %d, want 6", len(d2))
	}
	var got []byte
	got = append(got, d1...)
	got = append(got, d2...)
	if !bytes.Equal(got, data) {
		t.Fatalf("got %q, want %q", got, data)
	}
}

func TestRing512_PeekEmpty(t *testing.T) {
	var r ring512
	d1, d2 := r.Peek()
	if d1 != nil || d2 != nil {
		t.Fatalf("Peek on empty = (%v, %v), want (nil, nil)", d1, d2)
	}
}

func TestRing512_PeekTotalEqualsUsed(t *testing.T) {
	var r ring512
	// Test at many wrap positions.
	for offset := 0; offset < 512; offset += 37 {
		r.Reset()
		if offset > 0 {
			r.Put(make([]byte, offset))
			r.Discard(uint32(offset))
		}
		sz := 200
		r.Put(make([]byte, sz))
		d1, d2 := r.Peek()
		total := len(d1) + len(d2)
		if total != sz {
			t.Fatalf("offset=%d: Peek total=%d, want %d", offset, total, sz)
		}
	}
}

// --- Concurrent SPSC Test ---

func TestRing512_SPSC(t *testing.T) {
	for trial := 0; trial < 20; trial++ {
		var r ring512
		const totalBytes = 1 << 18
		produced := make([]byte, totalBytes)
		for i := range produced {
			produced[i] = byte(i + trial)
		}

		var wg sync.WaitGroup
		wg.Add(2)

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

		consumed := make([]byte, 0, totalBytes)
		go func() {
			defer wg.Done()
			for len(consumed) < totalBytes {
				d1, d2 := r.Peek()
				n := len(d1) + len(d2)
				if n == 0 {
					continue
				}
				consumed = append(consumed, d1...)
				consumed = append(consumed, d2...)
				r.Discard(uint32(n))
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
			d1, d2 := r.Peek()
			n := len(d1) + len(d2)
			if n == 0 {
				continue
			}
			consumed = append(consumed, d1...)
			consumed = append(consumed, d2...)
			r.Discard(uint32(n))
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

type refRing struct{ data []byte }

func (r *refRing) Put(d []byte) bool {
	if len(r.data)+len(d) > 512 {
		return false
	}
	r.data = append(r.data, d...)
	return true
}
func (r *refRing) Discard(n uint32) { r.data = r.data[n:] }
func (r *refRing) Used() uint32     { return uint32(len(r.data)) }

// FuzzRing512 compares ring512 against a trivially correct reference.
func FuzzRing512(f *testing.F) {
	f.Add([]byte{0, 10, 1, 5, 2, 0, 10, 1, 10})
	f.Add([]byte{0, 0})
	f.Add([]byte{0, 255, 0, 255, 1, 255, 1, 255})
	f.Add(bytes.Repeat([]byte{0, 64, 1, 64}, 50))
	f.Add([]byte{0, 200, 1, 100, 0, 156, 0, 156})

	f.Fuzz(func(t *testing.T, ops []byte) {
		var ring ring512
		var ref refRing

		i := 0
		for i+1 < len(ops) {
			op := ops[i] % 4
			arg := ops[i+1]
			i += 2

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
					t.Fatalf("Put(%d): ring=%v ref=%v", size, gotOK, refOK)
				}

			case 1: // Discard
				used := ring.Used()
				if used != ref.Used() {
					t.Fatalf("Used mismatch: ring=%d ref=%d", used, ref.Used())
				}
				if used == 0 {
					continue
				}
				n := uint32(arg) % (used + 1)
				ring.Discard(n)
				ref.Discard(n)

			case 2: // Peek + verify
				rUsed := ring.Used()
				if rUsed != ref.Used() {
					t.Fatalf("Used mismatch: ring=%d ref=%d", rUsed, ref.Used())
				}
				if rUsed == 0 {
					d1, d2 := ring.Peek()
					if d1 != nil || d2 != nil {
						t.Fatal("Peek non-nil on empty ring")
					}
					continue
				}
				got := peekAll(&ring)
				if uint32(len(got)) != rUsed {
					t.Fatalf("Peek returned %d bytes, Used=%d", len(got), rUsed)
				}
				if !bytes.Equal(got, ref.data) {
					t.Fatal("Peek data mismatch")
				}

			case 3: // Invariant
				if ring.Free()+ring.Used() != 512 {
					t.Fatalf("invariant: Free(%d)+Used(%d)!=512", ring.Free(), ring.Used())
				}
			}
		}

		if ring.Used() != ref.Used() {
			t.Fatalf("final Used mismatch: ring=%d ref=%d", ring.Used(), ref.Used())
		}
	})
}

// FuzzRing512_Op2 uses raw fuzz bytes as an operation stream with
// data integrity tracking.
func FuzzRing512_Op2(f *testing.F) {
	f.Add([]byte{0, 255, 0, 255, 0, 2, 1, 255})
	f.Add([]byte{0, 200, 1, 100, 0, 255, 0, 157, 1, 255, 1, 255})
	seed := make([]byte, 40)
	for i := range seed {
		if i%4 < 2 {
			seed[i] = 0
		} else {
			seed[i] = 1
		}
		if i%2 == 1 {
			seed[i] = byte(3 + i%13)
		}
	}
	f.Add(seed)

	f.Fuzz(func(t *testing.T, ops []byte) {
		const buflen = 512
		const maxOps = 128
		var ring ring512
		var written []byte
		totalRead := 0

		i := 0
		nops := 0
		for i+1 < len(ops) && nops < maxOps {
			op := ops[i] % 3
			sz := int(ops[i+1])
			i += 2
			nops++

			switch op {
			case 0: // Put
				if sz == 0 {
					continue
				}
				free := int(ring.Free())
				if sz > free {
					sz = free
				}
				if sz == 0 {
					continue
				}
				data := make([]byte, sz)
				for j := range data {
					data[j] = byte(len(written) + j)
				}
				if !ring.Put(data) {
					t.Fatalf("Put(%d) failed with Free()=%d", sz, free)
				}
				written = append(written, data...)

			case 1: // Read
				used := int(ring.Used())
				if used == 0 || sz == 0 {
					continue
				}
				if sz > used {
					sz = used
				}
				d1, d2 := ring.Peek()
				// Concatenate and take sz bytes.
				var got []byte
				if sz <= len(d1) {
					got = d1[:sz]
				} else {
					got = make([]byte, sz)
					copy(got, d1)
					copy(got[len(d1):], d2)
				}
				expect := written[totalRead : totalRead+sz]
				if !bytes.Equal(got, expect) {
					t.Fatalf("data mismatch at read offset %d", totalRead)
				}
				ring.Discard(uint32(sz))
				totalRead += sz

			case 2: // Reset
				ring.Reset()
				totalRead = len(written)
			}

			if ring.Free()+ring.Used() != buflen {
				t.Fatalf("invariant: Free(%d)+Used(%d)!=%d", ring.Free(), ring.Used(), buflen)
			}
			if int(ring.Used()) != len(written)-totalRead {
				t.Fatalf("Used()=%d expected %d", ring.Used(), len(written)-totalRead)
			}
		}

		// Final drain.
		d1, d2 := ring.Peek()
		var remaining []byte
		remaining = append(remaining, d1...)
		remaining = append(remaining, d2...)
		expect := written[totalRead:]
		if !bytes.Equal(remaining, expect) {
			t.Fatalf("final drain mismatch: got %d bytes, want %d", len(remaining), len(expect))
		}
	})
}
