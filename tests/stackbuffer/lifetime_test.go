//go:build tinygo

package stackbuffer

import (
	"bytes"
	"testing"
)

// These tests check that buffers that are live at the same time do not share
// stack space. All inputs come from the heap, so nothing is folded away.

// Sizes below, at and above the 32 byte buffer.
var lifetimeSizes = []int{0, 4, 31, 32, 33, 70}

//go:noinline
func hashStr(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h = (h ^ uint32(s[i])) * 16777619
	}
	return h ^ uint32(len(s))
}

//go:noinline
func hashBytes(b []byte) uint32 {
	h := uint32(2166136261)
	for _, c := range b {
		h = (h ^ uint32(c)) * 16777619
	}
	return h ^ uint32(len(b))
}

// input returns n bytes on the heap that depend on seed.
func input(n int, seed byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = seed + byte(i*7)
	}
	globalBytes = b
	return b
}

//go:noinline
func tail(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	return b[1:]
}

// Both strings are live when the second is created.
//
//go:noinline
func twoStringsLive(x, y []byte) (uint32, uint32) {
	a := string(x)
	b := string(y)
	return hashStr(a), hashStr(b)
}

// A string and a make are live together, and the make is written after the
// string exists.
//
//go:noinline
func stringAndMakeLive(x []byte, n int) (uint32, uint32) {
	a := string(x)
	m := make([]byte, size(n))
	for i := range m {
		m[i] = byte(i) + 3
	}
	return hashStr(a), hashBytes(m)
}

func TestTwoBuffersLiveTogether(t *testing.T) {
	for _, n := range lifetimeSizes {
		for _, m := range lifetimeSizes {
			x, y := input(n, 1), input(m, 100)
			want1, want2 := hashBytes(x), hashBytes(y)
			got1, got2 := twoStringsLive(x, y)
			if got1 != want1 || got2 != want2 {
				t.Errorf("strings %d,%d: got %x %x, want %x %x", n, m, got1, got2, want1, want2)
			}

			ref := make([]byte, m)
			for i := range ref {
				ref[i] = byte(i) + 3
			}
			globalBytes = ref
			got1, got2 = stringAndMakeLive(x, m)
			if got1 != want1 || got2 != hashBytes(ref) {
				t.Errorf("string and make %d,%d: got %x %x, want %x %x", n, m, got1, got2, want1, hashBytes(ref))
			}
		}
	}
}

// Each value is dead before the next is created, so they may share space. The
// results must still be right.
//
//go:noinline
func sequential(x, y, z []byte, n int) uint32 {
	h := hashStr(string(x))
	m := make([]byte, size(n))
	for i := range m {
		m[i] = byte(i) ^ 0x55
	}
	h = h*31 + hashBytes(m)
	h = h*31 + hashStr(string(y))
	h = h*31 + hashStr(string(z))
	return h
}

func TestSequentialBuffers(t *testing.T) {
	for _, n := range lifetimeSizes {
		x, y, z := input(n, 1), input(n+1, 50), input(n+2, 90)
		ref := make([]byte, n)
		for i := range ref {
			ref[i] = byte(i) ^ 0x55
		}
		globalBytes = ref
		want := hashBytes(x)
		want = want*31 + hashBytes(ref)
		want = want*31 + hashBytes(y)
		want = want*31 + hashBytes(z)
		if got := sequential(x, y, z, n); got != want {
			t.Errorf("size %d: got %x, want %x", n, got, want)
		}
	}
}

// a is only used in one arm of the branch, and b is created after the join.
//
//go:noinline
func usedInOneArm(x, y []byte, cond bool) uint32 {
	a := string(x)
	h := uint32(7)
	if cond {
		h = hashStr(a)
	}
	b := string(y)
	return h*31 + hashStr(b)
}

// a is live across an arm that creates b.
//
//go:noinline
func liveAcrossArm(x, y []byte, cond bool) uint32 {
	a := string(x)
	h := uint32(1)
	if cond {
		b := string(y)
		h += hashStr(b)
	}
	return h*31 + hashStr(a)
}

// Both arms make their own string, and a third is made after the join.
//
//go:noinline
func armsAndJoin(x, y, z []byte, cond bool) uint32 {
	var h uint32
	if cond {
		a := string(x)
		h = hashStr(a)
	} else {
		b := string(y)
		h = hashStr(b) + 1
	}
	c := string(z)
	return h*31 + hashStr(c)
}

func TestBranches(t *testing.T) {
	for _, n := range lifetimeSizes {
		for _, m := range lifetimeSizes {
			x, y, z := input(n, 1), input(m, 100), input(n+m, 200)
			for _, cond := range []bool{false, true} {
				h := uint32(7)
				if cond {
					h = hashBytes(x)
				}
				want := h*31 + hashBytes(y)
				if got := usedInOneArm(x, y, cond); got != want {
					t.Errorf("one arm %d,%d,%v: got %x, want %x", n, m, cond, got, want)
				}

				h = 1
				if cond {
					h += hashBytes(y)
				}
				want = h*31 + hashBytes(x)
				if got := liveAcrossArm(x, y, cond); got != want {
					t.Errorf("across arm %d,%d,%v: got %x, want %x", n, m, cond, got, want)
				}

				if cond {
					h = hashBytes(x)
				} else {
					h = hashBytes(y) + 1
				}
				want = h*31 + hashBytes(z)
				if got := armsAndJoin(x, y, z, cond); got != want {
					t.Errorf("arms and join %d,%d,%v: got %x, want %x", n, m, cond, got, want)
				}
			}
		}
	}
}

// prev is made before the loop and used in every iteration, while cur is made
// in each iteration.
//
//go:noinline
func liveAcrossLoop(x, y []byte, n int) uint32 {
	prev := string(x)
	var h uint32
	for i := 0; i < n; i++ {
		end := len(y)
		if end > 0 {
			end = 1 + i%len(y)
		}
		cur := string(y[:end])
		h = h*31 + hashStr(cur) + hashStr(prev)
	}
	return h
}

// The string of one iteration is still used in the next.
//
//go:noinline
func carriedInLoop(x []byte, n int) uint32 {
	prev := string(x)
	var h uint32
	for i := 0; i < n; i++ {
		cur := string(x[:len(x)-i%(len(x)+1)])
		h = h*31 + hashStr(prev)
		prev = cur
	}
	return h*31 + hashStr(prev)
}

func TestBuffersAcrossLoops(t *testing.T) {
	for _, n := range lifetimeSizes {
		for _, m := range lifetimeSizes {
			x, y := input(n, 1), input(m, 100)

			var want uint32
			for i := 0; i < 5; i++ {
				end := len(y)
				if end > 0 {
					end = 1 + i%len(y)
				}
				want = want*31 + hashBytes(y[:end]) + hashBytes(x)
			}
			if got := liveAcrossLoop(x, y, 5); got != want {
				t.Errorf("across loop %d,%d: got %x, want %x", n, m, got, want)
			}
		}

		x := input(n, 1)
		var want uint32
		prev := x
		for i := 0; i < 5; i++ {
			cur := x[:len(x)-i%(len(x)+1)]
			want = want*31 + hashBytes(prev)
			prev = cur
		}
		want = want*31 + hashBytes(prev)
		if got := carriedInLoop(x, 5); got != want {
			t.Errorf("carried %d: got %x, want %x", n, got, want)
		}
	}
}

// The result of a call can alias its argument, so the buffer lives as long as
// the result does.
//
//go:noinline
func aliasedResult(x, y []byte, n int) uint32 {
	m := make([]byte, size(n)+2)
	m[0] = ' '
	copy(m[1:], x)
	m[len(m)-1] = ' '
	trimmed := bytes.TrimSpace(m)
	rest := tail(m)
	other := string(y)
	h := hashStr(other)
	return h*31 + hashBytes(trimmed)*7 + hashBytes(rest)
}

func TestAliasedCallResult(t *testing.T) {
	for _, n := range lifetimeSizes {
		for _, m := range lifetimeSizes {
			x, y := input(n, 1), input(m, 100)
			ref := make([]byte, n+2)
			ref[0] = ' '
			copy(ref[1:], x)
			ref[len(ref)-1] = ' '
			globalBytes = ref
			want := hashBytes(y)*31 + hashBytes(bytes.TrimSpace(ref))*7 + hashBytes(tail(ref))
			if got := aliasedResult(x[:n], y, n); got != want {
				t.Errorf("sizes %d,%d: got %x, want %x", n, m, got, want)
			}
		}
	}
}

type holder struct {
	s string
	b []byte
}

// The pointers are kept in a local struct, so they live on through loads.
//
//go:noinline
func keptInStruct(x, y []byte, n int) uint32 {
	var h holder
	h.s = string(x)
	h.b = make([]byte, size(n))
	for i := range h.b {
		h.b[i] = byte(i) + 9
	}
	other := string(y)
	return hashStr(h.s)*31 + hashBytes(h.b)*7 + hashStr(other)
}

func TestPointersKeptInLocalStruct(t *testing.T) {
	for _, n := range lifetimeSizes {
		for _, m := range lifetimeSizes {
			x, y := input(n, 1), input(m, 100)
			ref := make([]byte, n)
			for i := range ref {
				ref[i] = byte(i) + 9
			}
			globalBytes = ref
			want := hashBytes(x)*31 + hashBytes(ref)*7 + hashBytes(y)
			if got := keptInStruct(x, y, n); got != want {
				t.Errorf("sizes %d,%d: got %x, want %x", n, m, got, want)
			}
		}
	}
}

// A fixed set of sites in one function, each a different size class, used in
// an order unrelated to creation order.
//
//go:noinline
func manyLive(a, b, c, d []byte) uint32 {
	sa, sb, sc, sd := string(a), string(b), string(c), string(d)
	h := hashStr(sc)
	h = h*31 + hashStr(sa)
	h = h*31 + hashStr(sd)
	h = h*31 + hashStr(sb)
	return h
}

func TestManyLive(t *testing.T) {
	for _, n := range lifetimeSizes {
		a, b, c, d := input(n, 1), input(32, 2), input(33, 3), input(n/2, 4)
		want := hashBytes(c)
		want = want*31 + hashBytes(a)
		want = want*31 + hashBytes(d)
		want = want*31 + hashBytes(b)
		if got := manyLive(a, b, c, d); got != want {
			t.Errorf("size %d: got %x, want %x", n, got, want)
		}
	}
}

// The size is the loop counter, and the allocation is in the block that jumps
// back to the loop header.
//
//go:noinline
func makeInLatch(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		b := make([]byte, i)
		for j := range b {
			b[j] = byte(j)
		}
		total += int(hashBytes(b))
	}
	return total
}

//go:noinline
func stringInLatch(x []byte) int {
	total := 0
	for i := 0; i < len(x); i++ {
		s := string(x[:i])
		total += int(hashStr(s))
	}
	return total
}

// The loop is one block, so the allocation is in the block that is its own
// successor and the size is a PHI of that block.
//
//go:noinline
func makeInSingleBlockLoop(n int) {
	for i := range n {
		s := make([]byte, i)
		_ = append(s, 42)
	}
}

//go:noinline
func stringInSingleBlockLoop(x []byte, n int) {
	for i := range n {
		s := string(x[:i])
		sinkInt += len(s)
	}
}

func TestSizeIsLoopCounter(t *testing.T) {
	makeInSingleBlockLoop(70)
	stringInSingleBlockLoop(input(70, 3), 70)

	want := 0
	for i := 0; i < 70; i++ {
		ref := make([]byte, i)
		for j := range ref {
			ref[j] = byte(j)
		}
		globalBytes = ref
		want += int(hashBytes(ref))
	}
	if got := makeInLatch(70); got != want {
		t.Errorf("make: got %x, want %x", got, want)
	}

	x := input(70, 5)
	want = 0
	for i := 0; i < len(x); i++ {
		want += int(hashBytes(x[:i]))
	}
	if got := stringInLatch(x); got != want {
		t.Errorf("string: got %x, want %x", got, want)
	}
}
